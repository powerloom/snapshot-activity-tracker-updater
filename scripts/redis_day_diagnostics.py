#!/usr/bin/env python3
"""
Read tracker Redis keys for one data market + day (local diagnostics).

  pip install redis
  REDIS_URL=redis://127.0.0.1:16379/0 \
    python scripts/redis_day_diagnostics.py 0xYourDataMarket 50

Compose: set REDIS_HOST_PORT (default 16379) and bind to localhost.
"""
from __future__ import annotations

import os
import sys


def main() -> None:
    if len(sys.argv) != 3:
        print("Usage: REDIS_URL=redis://host:port/db python scripts/redis_day_diagnostics.py <DATA_MARKET> <DAY>")
        sys.exit(1)
    dm_raw = sys.argv[1].strip()
    dm = dm_raw.lower()
    day = sys.argv[2].strip()
    url = os.environ.get("REDIS_URL", "redis://127.0.0.1:16379/0")
    try:
        import redis
    except ImportError:
        print("pip install redis")
        sys.exit(1)

    r = redis.Redis.from_url(url, decode_responses=True)

    quota_table = "DailySnapshotQuotaTableKey"
    q = r.hget(quota_table, dm_raw) or r.hget(quota_table, dm) or ""
    print("REDIS_URL", url)
    print("dailySnapshotQuota (hash)", quota_table, "->", repr(q) or "(missing)")

    elig_set = f"EligibleNodesByDay.{dm}.{day}"
    slots_set = f"SlotsWithSubmissionsByDay.{dm}.{day}"
    ne = r.scard(elig_set)
    ns = r.scard(slots_set)
    print("SCARD", elig_set, ne)
    print("SCARD", slots_set, ns)

    cur = r.get(f"CurrentDay.{dm}") or ""
    last = r.get(f"LastKnownDay.{dm}") or ""
    print("GET CurrentDay", repr(cur))
    print("GET LastKnownDay", repr(last))

    if ns and ns <= 20:
        for sid in sorted(r.smembers(slots_set), key=lambda x: int(x)):
            ev = r.get(f"EligibleSlotSubmissions.{dm}.{day}.{sid}") or r.get(
                f"SlotSubmissions.{dm}.{day}.{sid}"
            )
            print(f"  slot {sid} count {ev}")


if __name__ == "__main__":
    main()
