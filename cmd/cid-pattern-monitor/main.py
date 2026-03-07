#!/usr/bin/env python3
"""
CID Pattern Monitor — Analyze winning/losing CID patterns across epochs.

Data sources:
  1. Tracker-Updater dashboard API (port 8085):
     - /api/dashboard/summary → total epochs, slots, validators
     - /api/epochs → epoch list with summaries
     - /api/epochs/{id} → tally (winning CIDs, submission counts)
  2. DSV Node API (port 9091):
     - /api/v1/epochs/{id}/submissions → per-slot per-project CIDs + peer IDs

Detects:
  - Per-peer self-disagreement (same peer, same pool, same epoch, different CIDs)
  - CID dispersion per pool (unique CIDs / submissions — 1.0 = no consensus)
  - Per-slot loss streaks per pool
  - Multi-slot operator patterns (peer with many slots)

Outputs:
  - JSON report file (structured, for feeding back into analysis)
  - Human-readable summary to stdout

Usage:
  # Analyze last N epochs
  python3 main.py --last 50

  # Analyze specific epoch range
  python3 main.py --from-epoch 24606500 --to-epoch 24606600

  # Continuous monitoring
  python3 main.py --watch

Environment variables (override CLI args):
  TRACKER_URL           Tracker-updater dashboard API (default: http://localhost:8085)
  DSV_NODE_URL          DSV node submissions API (default: http://localhost:9091)
  REPORT_DIR            Where to write reports (default: ./reports)
"""

import argparse
import json
import os
import sys
import time
from collections import defaultdict
from dataclasses import dataclass, field, asdict
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional, Set, Tuple

try:
    import httpx

    def http_get(url: str, timeout: int = 15) -> Optional[dict]:
        try:
            with httpx.Client(timeout=timeout) as client:
                resp = client.get(url, headers={"Accept": "application/json"})
                resp.raise_for_status()
                return resp.json()
        except Exception as e:
            print(f"  [WARN] GET {url}: {e}", file=sys.stderr)
            return None
except ImportError:
    import urllib.request
    import urllib.error

    def http_get(url: str, timeout: int = 15) -> Optional[dict]:
        try:
            req = urllib.request.Request(url, headers={"Accept": "application/json"})
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                return json.loads(resp.read().decode())
        except Exception as e:
            print(f"  [WARN] GET {url}: {e}", file=sys.stderr)
            return None


# ---------------------------------------------------------------------------
# Data fetching
# ---------------------------------------------------------------------------

def fetch_dashboard_summary(tracker_url: str) -> Optional[dict]:
    """GET /api/dashboard/summary → {total_epochs, total_slots, ...}."""
    return http_get(f"{tracker_url}/api/dashboard/summary")


def fetch_epoch_list(tracker_url: str) -> Optional[List[dict]]:
    """GET /api/epochs → {epochs: [{epoch_id, timestamp, ...}, ...]}."""
    data = http_get(f"{tracker_url}/api/epochs")
    if data and "epochs" in data:
        return data["epochs"]
    return None


def fetch_epoch_detail(tracker_url: str, epoch_id: int) -> Optional[dict]:
    """GET /api/epochs/{id} → tally (winning CIDs, submission_counts)."""
    return http_get(f"{tracker_url}/api/epochs/{epoch_id}")


def fetch_submissions(node_url: str, epoch_id: int) -> Optional[List[dict]]:
    """GET /api/v1/epochs/{id}/submissions → per-slot per-project CIDs."""
    data = http_get(f"{node_url}/api/v1/epochs/{epoch_id}/submissions")
    if data and "submissions" in data:
        return data["submissions"]
    return None


# ---------------------------------------------------------------------------
# Analysis types
# ---------------------------------------------------------------------------

@dataclass
class PeerProfile:
    peer_id: str
    slot_ids: Set[str] = field(default_factory=set)
    total_submissions: int = 0
    wins: int = 0
    losses: int = 0
    self_disagreements: int = 0  # epochs where this peer submitted different CIDs for same pool

    @property
    def win_rate(self) -> float:
        total = self.wins + self.losses
        return self.wins / total if total > 0 else 0.0

    @property
    def is_multi_slot(self) -> bool:
        return len(self.slot_ids) > 1


@dataclass
class PoolEpochResult:
    pool: str
    epoch_id: int
    total_submissions: int
    unique_cids: int
    winning_cid: str
    winning_count: int
    dispersion: float  # unique_cids / total_submissions
    peer_self_disagreements: List[dict] = field(default_factory=list)
    slot_results: List[dict] = field(default_factory=list)


@dataclass
class SlotPoolStreak:
    slot_id: str
    pool: str
    current_streak: int = 0  # positive=wins, negative=losses
    max_loss_streak: int = 0
    wins: int = 0
    losses: int = 0
    last_epoch: int = 0


# ---------------------------------------------------------------------------
# Core analysis
# ---------------------------------------------------------------------------

class CIDAnalyzer:
    def __init__(self):
        self.peer_profiles: Dict[str, PeerProfile] = defaultdict(
            lambda: PeerProfile(peer_id="")
        )
        self.slot_pool_streaks: Dict[str, SlotPoolStreak] = {}
        self.pool_dispersion_history: Dict[str, List[float]] = defaultdict(list)
        self.epoch_results: List[dict] = []
        self.alerts: List[dict] = []
        self.processed_epochs: List[int] = []

    def analyze_epoch(
        self,
        epoch_id: int,
        tally: dict,
        submissions: List[dict],
    ) -> dict:
        """Analyze a single epoch. Returns epoch report dict."""
        winning_cids = tally.get("aggregated_projects", {})
        eligible_nodes = tally.get("eligible_nodes_count", 0)

        # Group submissions by project
        by_project: Dict[str, List[dict]] = defaultdict(list)
        for sub in submissions:
            by_project[sub["project_id"]].append(sub)

        pool_results = []

        for project_id, subs in by_project.items():
            winning_cid = winning_cids.get(project_id)
            if not winning_cid:
                continue

            parts = project_id.split(":")
            pool = parts[1] if len(parts) >= 2 else project_id

            # CID distribution
            cid_slots: Dict[str, List[str]] = defaultdict(list)
            for sub in subs:
                cid_slots[sub["snapshot_cid"]].append(sub["slot_id"])

            total = len(subs)
            unique = len(cid_slots)
            winning_count = len(cid_slots.get(winning_cid, []))
            dispersion = unique / total if total > 0 else 0.0

            # Per-peer analysis: detect same-peer self-disagreement
            peer_cids: Dict[str, Dict[str, List[str]]] = defaultdict(
                lambda: defaultdict(list)
            )
            for sub in subs:
                peer_cids[sub["peer_id"]][sub["snapshot_cid"]].append(sub["slot_id"])

            peer_disagreements = []
            for peer_id, cids_map in peer_cids.items():
                if len(cids_map) > 1:
                    peer_disagreements.append({
                        "peer_id": peer_id,
                        "unique_cids": len(cids_map),
                        "slots_per_cid": {
                            cid[:24]: slots for cid, slots in cids_map.items()
                        },
                    })
                    profile = self.peer_profiles[peer_id]
                    profile.self_disagreements += 1

            # Update peer profiles
            for sub in subs:
                pid = sub["peer_id"]
                profile = self.peer_profiles[pid]
                profile.peer_id = pid
                profile.slot_ids.add(sub["slot_id"])
                profile.total_submissions += 1
                won = sub["snapshot_cid"] == winning_cid
                if won:
                    profile.wins += 1
                else:
                    profile.losses += 1

            # Slot results + streak tracking
            slot_results = []
            for sub in subs:
                sid = sub["slot_id"]
                won = sub["snapshot_cid"] == winning_cid
                key = f"{sid}:{pool}"

                if key not in self.slot_pool_streaks:
                    self.slot_pool_streaks[key] = SlotPoolStreak(
                        slot_id=sid, pool=pool
                    )
                streak = self.slot_pool_streaks[key]

                if won:
                    streak.wins += 1
                    streak.current_streak = max(streak.current_streak, 0) + 1
                else:
                    streak.losses += 1
                    streak.current_streak = min(streak.current_streak, 0) - 1
                    streak.max_loss_streak = max(
                        streak.max_loss_streak, abs(streak.current_streak)
                    )
                    if abs(streak.current_streak) >= 5:
                        self.alerts.append({
                            "type": "loss_streak",
                            "epoch_id": epoch_id,
                            "slot_id": sid,
                            "pool": pool,
                            "peer_id": sub["peer_id"],
                            "streak": abs(streak.current_streak),
                        })

                streak.last_epoch = epoch_id
                slot_results.append({
                    "slot_id": sid,
                    "peer_id": sub["peer_id"],
                    "cid": sub["snapshot_cid"],
                    "won": won,
                })

            self.pool_dispersion_history[pool].append(dispersion)

            pool_results.append({
                "pool": pool,
                "project_id": project_id,
                "total_submissions": total,
                "unique_cids": unique,
                "winning_cid": winning_cid,
                "winning_count": winning_count,
                "dispersion": round(dispersion, 3),
                "agreement_pct": round(winning_count / total * 100, 1) if total else 0,
                "peer_self_disagreements": peer_disagreements,
                "cid_distribution": {
                    cid[:24] + "...": {
                        "count": len(slots),
                        "slots": slots,
                    }
                    for cid, slots in cid_slots.items()
                },
                "slot_results": slot_results,
            })

        # Sort by dispersion (most chaotic first)
        pool_results.sort(key=lambda r: r["dispersion"], reverse=True)

        epoch_report = {
            "epoch_id": epoch_id,
            "timestamp": tally.get("timestamp"),
            "eligible_nodes": eligible_nodes,
            "total_pools": len(pool_results),
            "pools_with_disagreement": sum(
                1 for r in pool_results if r["unique_cids"] > 1
            ),
            "pools_with_self_disagree": sum(
                1 for r in pool_results if r["peer_self_disagreements"]
            ),
            "pool_results": pool_results,
        }

        self.epoch_results.append(epoch_report)
        self.processed_epochs.append(epoch_id)
        return epoch_report

    def generate_summary(self) -> dict:
        """Generate cross-epoch summary report."""
        # Pool dispersion rankings
        pool_avg_dispersion = {}
        for pool, dispersions in self.pool_dispersion_history.items():
            pool_avg_dispersion[pool] = {
                "avg_dispersion": round(sum(dispersions) / len(dispersions), 3),
                "max_dispersion": round(max(dispersions), 3),
                "epochs_analyzed": len(dispersions),
                "epochs_with_full_chaos": sum(1 for d in dispersions if d >= 0.8),
                "epochs_with_perfect_consensus": sum(
                    1 for d in dispersions if d <= 0.2
                ),
            }
        sorted_pools = sorted(
            pool_avg_dispersion.items(),
            key=lambda x: x[1]["avg_dispersion"],
            reverse=True,
        )

        # Peer profiles
        peer_summaries = []
        for pid, profile in sorted(
            self.peer_profiles.items(),
            key=lambda x: len(x[1].slot_ids),
            reverse=True,
        ):
            peer_summaries.append({
                "peer_id": pid,
                "slot_count": len(profile.slot_ids),
                "is_multi_slot_operator": profile.is_multi_slot,
                "total_submissions": profile.total_submissions,
                "wins": profile.wins,
                "losses": profile.losses,
                "win_rate_pct": round(profile.win_rate * 100, 1),
                "self_disagreement_epochs": profile.self_disagreements,
            })

        # Top loss streaks
        loss_streaks = []
        for key, streak in sorted(
            self.slot_pool_streaks.items(),
            key=lambda x: x[1].max_loss_streak,
            reverse=True,
        ):
            if streak.max_loss_streak >= 3:
                total = streak.wins + streak.losses
                loss_streaks.append({
                    "slot_id": streak.slot_id,
                    "pool": streak.pool,
                    "max_loss_streak": streak.max_loss_streak,
                    "current_streak": streak.current_streak,
                    "win_rate_pct": round(
                        streak.wins / total * 100, 1
                    ) if total else 0,
                    "total_epochs": total,
                })

        return {
            "report_generated": datetime.utcnow().isoformat() + "Z",
            "epochs_analyzed": len(self.processed_epochs),
            "epoch_range": {
                "first": min(self.processed_epochs) if self.processed_epochs else None,
                "last": max(self.processed_epochs) if self.processed_epochs else None,
            },
            "pool_dispersion_rankings": sorted_pools,
            "peer_profiles": peer_summaries,
            "top_loss_streaks": loss_streaks[:30],
            "alerts": self.alerts[-50:],
        }


# ---------------------------------------------------------------------------
# Report output
# ---------------------------------------------------------------------------

def print_epoch_summary(report: dict):
    """Print concise epoch summary to stdout."""
    eid = report["epoch_id"]
    total = report["total_pools"]
    disagree = report["pools_with_disagreement"]
    self_disagree = report["pools_with_self_disagree"]

    if disagree == 0:
        print(f"  epoch {eid}: {total} pools, all consensus OK")
        return

    print(f"\n  epoch {eid}: {total} pools, {disagree} with CID disagreement, "
          f"{self_disagree} with same-peer self-disagreement")

    for r in report["pool_results"]:
        if r["unique_cids"] <= 1:
            continue
        flags = []
        if r["peer_self_disagreements"]:
            peers = [d["peer_id"][:16] + ".." for d in r["peer_self_disagreements"]]
            flags.append(f"SELF-DISAGREE({','.join(peers)})")
        if r["dispersion"] >= 0.8:
            flags.append("HIGH-DISPERSION")
        if r["winning_count"] == 1 and r["total_submissions"] > 1:
            flags.append("MINORITY-WINNER")

        flag_str = " " + " ".join(flags) if flags else ""
        print(
            f"    {r['pool'][:18]} | {r['total_submissions']} subs, "
            f"{r['unique_cids']} CIDs, {r['agreement_pct']}% agree{flag_str}"
        )


def print_final_summary(summary: dict):
    """Print human-readable final summary."""
    print(f"\n{'='*80}")
    print(f"ANALYSIS SUMMARY")
    print(f"Epochs: {summary['epoch_range']['first']} → {summary['epoch_range']['last']} "
          f"({summary['epochs_analyzed']} analyzed)")
    print(f"{'='*80}")

    print(f"\n--- POOL CID DISPERSION (worst first) ---")
    for pool, stats in summary["pool_dispersion_rankings"][:15]:
        print(
            f"  {pool[:18]} | avg={stats['avg_dispersion']:.3f} "
            f"max={stats['max_dispersion']:.3f} | "
            f"{stats['epochs_with_full_chaos']} chaotic / "
            f"{stats['epochs_with_perfect_consensus']} perfect / "
            f"{stats['epochs_analyzed']} total"
        )

    print(f"\n--- PEER PROFILES ---")
    for p in summary["peer_profiles"]:
        label = "MULTI-SLOT" if p["is_multi_slot_operator"] else "single"
        disagree_note = ""
        if p["self_disagreement_epochs"] > 0:
            disagree_note = f" !! {p['self_disagreement_epochs']} self-disagree epochs"
        print(
            f"  {p['peer_id'][:24]}.. | {p['slot_count']:>3} slots | "
            f"win={p['win_rate_pct']:5.1f}% ({p['wins']}/{p['total_submissions']}) "
            f"[{label}]{disagree_note}"
        )

    if summary["top_loss_streaks"]:
        print(f"\n--- TOP LOSS STREAKS (>= 3 consecutive) ---")
        for s in summary["top_loss_streaks"][:15]:
            print(
                f"  slot {s['slot_id']:>5} | {s['pool'][:18]} | "
                f"max_streak={s['max_loss_streak']}, current={s['current_streak']}, "
                f"win_rate={s['win_rate_pct']}%"
            )

    if summary["alerts"]:
        print(f"\n--- ALERTS ({len(summary['alerts'])}) ---")
        for a in summary["alerts"][-10:]:
            print(f"  [{a['type']}] epoch={a['epoch_id']} slot={a['slot_id']} "
                  f"pool={a['pool'][:18]} streak={a['streak']}")


def write_report(report_dir: Path, report: dict, filename: str):
    """Write JSON report to file."""
    report_dir.mkdir(parents=True, exist_ok=True)
    path = report_dir / filename
    with open(path, "w") as f:
        json.dump(report, f, indent=2, default=str)
    print(f"  Report written: {path}")


# ---------------------------------------------------------------------------
# Main modes
# ---------------------------------------------------------------------------

def run_batch(
    tracker_url: str,
    node_url: str,
    report_dir: Path,
    from_epoch: Optional[int],
    to_epoch: Optional[int],
    last_n: Optional[int],
):
    """Analyze a batch of epochs."""
    # Show dashboard summary
    summary_data = fetch_dashboard_summary(tracker_url)
    if summary_data:
        print(f"  Dashboard: {summary_data.get('total_epochs', '?')} epochs, "
              f"{summary_data.get('total_slots', '?')} slots, "
              f"{summary_data.get('total_validators', '?')} validators, "
              f"day {summary_data.get('current_day', '?')}")

    print("Fetching epoch list...")
    epochs = fetch_epoch_list(tracker_url)
    if not epochs:
        print("ERROR: Could not fetch epoch list from tracker API", file=sys.stderr)
        sys.exit(1)

    epoch_ids = sorted(e["epoch_id"] for e in epochs)
    print(f"  Available: {len(epoch_ids)} epochs ({epoch_ids[0]}..{epoch_ids[-1]})")

    if last_n:
        epoch_ids = epoch_ids[-last_n:]
    elif from_epoch or to_epoch:
        lo = from_epoch or epoch_ids[0]
        hi = to_epoch or epoch_ids[-1]
        epoch_ids = [e for e in epoch_ids if lo <= e <= hi]

    print(f"  Analyzing: {len(epoch_ids)} epochs ({epoch_ids[0]}..{epoch_ids[-1]})")

    analyzer = CIDAnalyzer()
    failed = 0

    for i, eid in enumerate(epoch_ids):
        tally = fetch_epoch_detail(tracker_url, eid)
        if not tally:
            failed += 1
            continue

        submissions = fetch_submissions(node_url, eid)
        if not submissions:
            failed += 1
            continue

        report = analyzer.analyze_epoch(eid, tally, submissions)
        print_epoch_summary(report)

        if (i + 1) % 10 == 0:
            print(f"  ... processed {i+1}/{len(epoch_ids)}")

    if failed:
        print(f"\n  [WARN] Failed to fetch data for {failed}/{len(epoch_ids)} epochs")

    # Generate and write reports
    summary = analyzer.generate_summary()

    ts = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
    epoch_range = f"{epoch_ids[0]}_{epoch_ids[-1]}" if epoch_ids else "none"

    write_report(
        report_dir,
        summary,
        f"summary_{epoch_range}_{ts}.json",
    )
    write_report(
        report_dir,
        {"epochs": analyzer.epoch_results},
        f"epochs_{epoch_range}_{ts}.json",
    )

    print_final_summary(summary)


def run_watch(
    tracker_url: str,
    node_url: str,
    report_dir: Path,
    interval: int,
    report_every: int,
):
    """Continuous monitoring mode."""
    print(f"Watch mode: polling every {interval}s, report every {report_every} epochs")

    analyzer = CIDAnalyzer()
    seen: Set[int] = set()
    epochs_since_report = 0

    while True:
        epochs = fetch_epoch_list(tracker_url)
        if not epochs:
            time.sleep(interval)
            continue

        new_ids = sorted(e["epoch_id"] for e in epochs if e["epoch_id"] not in seen)

        for eid in new_ids:
            tally = fetch_epoch_detail(tracker_url, eid)
            submissions = fetch_submissions(node_url, eid) if tally else None

            if tally and submissions:
                report = analyzer.analyze_epoch(eid, tally, submissions)
                print_epoch_summary(report)
                epochs_since_report += 1

            seen.add(eid)

            if epochs_since_report >= report_every:
                summary = analyzer.generate_summary()
                ts = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
                write_report(report_dir, summary, f"summary_watch_{ts}.json")
                print_final_summary(summary)
                epochs_since_report = 0

        time.sleep(interval)


def main():
    parser = argparse.ArgumentParser(
        description="CID Pattern Monitor — Analyze snapshot CID consensus patterns",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python3 main.py --last 50
  python3 main.py --from-epoch 24606500 --to-epoch 24606600
  python3 main.py --watch
        """,
    )
    parser.add_argument(
        "--tracker-url",
        default=os.environ.get("TRACKER_URL", "http://localhost:8085"),
        help="Tracker-updater dashboard API base URL",
    )
    parser.add_argument(
        "--node-url",
        default=os.environ.get("DSV_NODE_URL", "http://localhost:9091"),
        help="DSV node submissions API base URL",
    )
    parser.add_argument(
        "--report-dir",
        default=os.environ.get("REPORT_DIR", "./reports"),
        help="Directory for JSON report output",
    )
    parser.add_argument("--last", type=int, help="Analyze last N epochs")
    parser.add_argument("--from-epoch", type=int, help="Start epoch (inclusive)")
    parser.add_argument("--to-epoch", type=int, help="End epoch (inclusive)")
    parser.add_argument(
        "--watch", action="store_true", help="Continuous monitoring mode"
    )
    parser.add_argument(
        "--interval",
        type=int,
        default=int(os.environ.get("MONITOR_INTERVAL", "15")),
        help="Poll interval in seconds (watch mode)",
    )
    parser.add_argument(
        "--report-every",
        type=int,
        default=int(os.environ.get("REPORT_INTERVAL", "100")),
        help="Epochs between summary reports (watch mode)",
    )

    args = parser.parse_args()
    report_dir = Path(args.report_dir)

    print(f"CID Pattern Monitor")
    print(f"  tracker_url: {args.tracker_url}")
    print(f"  node_url:    {args.node_url}")
    print(f"  report_dir:  {report_dir}")

    if args.watch:
        try:
            run_watch(
                args.tracker_url, args.node_url, report_dir,
                args.interval, args.report_every,
            )
        except KeyboardInterrupt:
            print("\nStopped.")
    else:
        run_batch(
            args.tracker_url, args.node_url, report_dir,
            args.from_epoch, args.to_epoch, args.last,
        )


if __name__ == "__main__":
    main()
