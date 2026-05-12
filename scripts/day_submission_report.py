#!/usr/bin/env python3
"""
Summarize per-slot submission counts and related Redis state for one protocol day.

Defaults to day 51. Keys match Go redis/keys.go (lowercased data market).

  pip install redis
  REDIS_URL=redis://127.0.0.1:16379/0 \\
    python scripts/day_submission_report.py 0xYourDataMarket
  python scripts/day_submission_report.py 0xYourDataMarket 51 --print-slots 30

Env:
  REDIS_URL   default redis://127.0.0.1:16379/0
"""
from __future__ import annotations

import argparse
import os
import statistics
import sys
from typing import Iterable


DEFAULT_DAY = "51"
QUOTA_HASH = "DailySnapshotQuotaTableKey"


def chunked(xs: list[str], n: int) -> Iterable[list[str]]:
    for i in range(0, len(xs), n):
        yield xs[i : i + n]


def scan_match(r, pattern: str) -> list[str]:
    out: list[str] = []
    cursor = 0
    while True:
        cursor, keys = r.scan(cursor=cursor, match=pattern, count=200)
        out.extend(keys)
        if cursor == 0:
            break
    return sorted(set(out))


def main() -> None:
    p = argparse.ArgumentParser(description="Day N submission counts from tracker Redis")
    p.add_argument("data_market", help="Data market address (0x...)")
    p.add_argument(
        "day",
        nargs="?",
        default=DEFAULT_DAY,
        help=f"Protocol day string (default {DEFAULT_DAY})",
    )
    p.add_argument(
        "--print-slots",
        type=int,
        default=0,
        metavar="N",
        help="Print first N slots with counts (0 = summary only)",
    )
    p.add_argument(
        "--epoch-hashes",
        action="store_true",
        help="List EligibleSlotSubmissionsByEpoch.<dm>.<day>.* keys (SCAN)",
    )
    args = p.parse_args()

    dm_raw = args.data_market.strip()
    dm = dm_raw.lower()
    day = str(args.day).strip()

    try:
        import redis
    except ImportError:
        print("pip install redis", file=sys.stderr)
        sys.exit(1)

    url = os.environ.get("REDIS_URL", "redis://127.0.0.1:16379/0")
    r = redis.Redis.from_url(url, decode_responses=True)

    q_raw = r.hget(QUOTA_HASH, dm_raw) or r.hget(QUOTA_HASH, dm) or ""
    try:
        quota = int(q_raw) if q_raw != "" else 0
    except ValueError:
        quota = 0

    elig_set = f"EligibleNodesByDay.{dm}.{day}"
    slots_set = f"SlotsWithSubmissionsByDay.{dm}.{day}"
    ne = r.scard(elig_set)
    ns = r.scard(slots_set)

    cur = r.get(f"CurrentDay.{dm}") or ""
    last = r.get(f"LastKnownDay.{dm}") or ""

    print("REDIS_URL", url)
    print("data_market", dm_raw, "(keys:", dm + ")")
    print("day", day)
    print("dailySnapshotQuota", QUOTA_HASH, "->", repr(q_raw) or "(missing)", f"(parsed={quota})")
    print("GET CurrentDay", repr(cur))
    print("GET LastKnownDay", repr(last))
    print("SCARD", elig_set, ne)
    print("SCARD", slots_set, ns)

    # Same fallback as GetCountsForDay: prefer all slots with submissions
    slot_ids = sorted(r.smembers(slots_set), key=lambda x: int(x)) if ns else []
    if not slot_ids and ne:
        slot_ids = sorted(r.smembers(elig_set), key=lambda x: int(x))

    counts: dict[int, int] = {}
    for batch in chunked(slot_ids, 400):
        pipe = r.pipeline(transaction=False)
        for sid in batch:
            pipe.get(f"EligibleSlotSubmissions.{dm}.{day}.{sid}")
        elig_vals = pipe.execute()
        fallback_sids = [sid for sid, ev in zip(batch, elig_vals) if ev in (None, "")]
        slot_map: dict[str, str | None] = {}
        if fallback_sids:
            pipe = r.pipeline(transaction=False)
            for sid in fallback_sids:
                pipe.get(f"SlotSubmissions.{dm}.{day}.{sid}")
            for sid, sv in zip(fallback_sids, pipe.execute()):
                slot_map[sid] = sv
        for sid, ev in zip(batch, elig_vals):
            raw = ev if ev not in (None, "") else slot_map.get(sid)
            if raw in (None, ""):
                continue
            try:
                counts[int(sid)] = int(raw)
            except ValueError:
                counts[int(sid)] = -1

    if not counts and not slot_ids:
        print("No slot members in Sets; nothing to aggregate.")
    elif not counts and slot_ids:
        print(f"Warning: {len(slot_ids)} set members but no string counts found (keys missing?)")

    if counts:
        vals = sorted(counts.values())
        total_slots = len(counts)
        sum_counts = sum(v for v in vals if v >= 0)
        pos = [v for v in vals if v >= 0]
        ge_q = sum(1 for v in pos if (quota == 0 and v > 0) or (quota > 0 and v >= quota))
        print("--- submission counts (from EligibleSlotSubmissions*, else SlotSubmissions*) ---")
        print("slots_with_numeric_count", total_slots)
        print("sum_of_counts", sum_counts)
        if pos:
            print("min", min(pos), "max", max(pos), "mean", f"{statistics.mean(pos):.2f}")
            print("median", statistics.median(pos))
        print(
            "slots_meeting_quota",
            ge_q,
            f"(quota rule: {'>0' if quota == 0 else f'>={quota}'})",
        )
        in_elig_not_in_map = set(int(x) for x in r.smembers(elig_set)) - set(counts.keys())
        if in_elig_not_in_map:
            print("note: EligibleNodesByDay members missing count keys:", len(in_elig_not_in_map))

    if args.print_slots > 0 and counts:
        print(f"--- first {args.print_slots} slots by slot id ---")
        for sid in sorted(counts.keys())[: args.print_slots]:
            print(f"  slot {sid}\tcount {counts[sid]}")

    if args.epoch_hashes:
        pat = f"EligibleSlotSubmissionsByEpoch.{dm}.{day}.*"
        keys = scan_match(r, pat)
        print("--- SCAN", pat, f"-> {len(keys)} keys ---")
        for k in keys[:50]:
            n = r.hlen(k)
            print(f"  {k}  hlen={n}")
        if len(keys) > 50:
            print(f"  ... {len(keys) - 50} more")


if __name__ == "__main__":
    main()
