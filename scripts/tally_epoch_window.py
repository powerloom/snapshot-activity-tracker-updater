#!/usr/bin/env python3
"""
Load per-epoch tally JSON from Redis (TallyEpoch.<lowercase_dm>.<epoch>) for a window
of epochs — e.g. last ~10 epochs before a day rollover — to compare slot counts vs a quota.

  pip install redis
  REDIS_URL=redis://127.0.0.1:16379/0 python scripts/tally_epoch_window.py \\
    0xYourDM 24285100 24285115 400

Prints per epoch: slot count, distribution buckets vs quota (ge / in [q-100,q) / lt q-100),
and writes one JSONL file if OUT=path is set.
"""
from __future__ import annotations

import json
import os
import sys


def bucket_counts(submission_counts: dict, q: int) -> dict:
    ge = hi = lo = 0
    for _, v in submission_counts.items():
        try:
            c = int(v)
        except (TypeError, ValueError):
            continue
        if c >= q:
            ge += 1
        elif c >= max(0, q - 100):
            hi += 1
        else:
            lo += 1
    return {"slots_ge_quota": ge, f"slots_in_[{max(0,q-100)},{q})": hi, f"slots_lt_{max(0,q-100)}": lo}


def main() -> None:
    if len(sys.argv) not in (4, 5):
        print(
            "Usage: REDIS_URL=redis://... python tally_epoch_window.py <DATA_MARKET> "
            "<EPOCH_START> <EPOCH_END> [QUOTA_FOR_BUCKETS]"
        )
        sys.exit(1)
    dm = sys.argv[1].strip().lower()
    e0 = int(sys.argv[2])
    e1 = int(sys.argv[3])
    quota = int(sys.argv[4]) if len(sys.argv) == 5 else 0
    if e0 > e1:
        e0, e1 = e1, e0

    try:
        import redis
    except ImportError:
        print("pip install redis")
        sys.exit(1)

    r = redis.Redis.from_url(os.environ.get("REDIS_URL", "redis://127.0.0.1:16379/0"), decode_responses=True)
    out_path = os.environ.get("OUT")
    jsonl_lines: list[str] = []

    for e in range(e0, e1 + 1):
        key = f"TallyEpoch.{dm}.{e}"
        raw = r.get(key)
        if not raw:
            print(f"epoch {e} MISSING ({key})")
            continue
        doc = json.loads(raw)
        sc = doc.get("submission_counts") or {}
        row = {
            "epoch_id": doc.get("epoch_id", e),
            "timestamp": doc.get("timestamp"),
            "n_slots_in_tally": len(sc),
            "eligible_nodes_count": doc.get("eligible_nodes_count"),
        }
        if quota > 0:
            row["buckets_vs_quota"] = bucket_counts(sc, quota)
        print(json.dumps(row))
        jsonl_lines.append(raw)

    if out_path:
        with open(out_path, "w", encoding="utf-8") as f:
            for line in jsonl_lines:
                f.write(line + "\n")
        print("wrote", out_path, file=sys.stderr)


if __name__ == "__main__":
    main()
