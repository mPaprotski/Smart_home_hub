"""Small strict wire contract for derived scalar observations, not source media."""

import math
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import Literal, cast

Kind = Literal["illuminance", "temperature", "audio_rms"]
_UNITS: dict[str, str] = {"illuminance": "lx", "temperature": "°C", "audio_rms": "ratio"}
_KEYS = frozenset(
    {"schema_version", "event_id", "producer", "kind", "value", "unit", "observed_at"}
)


@dataclass(frozen=True, slots=True)
class Observation:
    event_id: str
    producer: str
    kind: Kind
    value: float
    unit: str
    observed_at: datetime
    schema_version: int = 1

    @classmethod
    def parse(cls, payload: dict[str, object]) -> "Observation":
        if payload.keys() != _KEYS or type(payload["schema_version"]) is not int:
            raise ValueError("invalid observation fields")
        if payload["schema_version"] != 1:
            raise ValueError("unsupported schema version")
        event_id, producer = payload["event_id"], payload["producer"]
        if not all(isinstance(s, str) and 0 < len(s) <= 128 for s in (event_id, producer)):
            raise ValueError("invalid observation identity")
        kind, unit = payload["kind"], payload["unit"]
        if not isinstance(kind, str) or kind not in _UNITS or unit != _UNITS[kind]:
            raise ValueError("invalid observation kind or unit")
        value = payload["value"]
        if type(value) not in (int, float):
            raise ValueError("value must be a finite number")
        try:
            number = float(cast("int | float", value))
        except OverflowError as exc:
            raise ValueError("value outside numeric range") from exc
        if not math.isfinite(number):
            raise ValueError("value must be finite")
        if kind in ("illuminance", "audio_rms") and number < 0:
            raise ValueError("negative observation")
        if kind == "audio_rms" and number > 1:
            raise ValueError("RMS outside normalized range")
        raw_time = payload["observed_at"]
        if not isinstance(raw_time, str) or len(raw_time) > 40:
            raise ValueError("invalid observation timestamp")
        try:
            observed_at = datetime.fromisoformat(raw_time)
        except ValueError as exc:
            raise ValueError("invalid observation timestamp") from exc
        if observed_at.utcoffset() != timedelta(0):
            raise ValueError("timestamp must be UTC with timezone")
        return cls(
            cast(str, event_id),
            cast(str, producer),
            cast(Kind, kind),
            number,
            unit,
            observed_at,
        )
