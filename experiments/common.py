"""Helpers for bounded, privacy-preserving feasibility probes (not production Core)."""
import json
import math
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def emit(probe, status, checks, **details):
    print(json.dumps({"probe": probe, "status": status,
                      "executed_at_utc": datetime.now(timezone.utc).isoformat(),
                      "client_date": "2026-10-25", "checks": checks, **details},
                     ensure_ascii=False, indent=2, allow_nan=False))


def require_finite(value, name, low=None, high=None):
    if isinstance(value, bool) or not isinstance(value, (int, float)) or not math.isfinite(value):
        raise ValueError(f"{name}: finite number required")
    if low is not None and value < low or high is not None and value > high:
        raise ValueError(f"{name}: out of range")
    return value
