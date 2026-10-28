import math
from dataclasses import FrozenInstanceError

import pytest

from smarthome_ml.observations import Observation


def valid():
    return dict(
        schema_version=1,
        event_id="synthetic-1",
        producer="test",
        kind="illuminance",
        value=12.5,
        unit="lx",
        observed_at="2026-10-25T12:00:00Z",
    )


def test_valid_owned_immutable_observation():
    data = valid()
    result = Observation.parse(data)
    data["value"] = 99
    assert result.value == 12.5
    with pytest.raises(FrozenInstanceError):
        result.value = 99


@pytest.mark.parametrize(
    "field,value",
    [
        ("schema_version", True),
        ("schema_version", 2),
        ("value", True),
        ("value", math.nan),
        ("value", math.inf),
        ("value", -1),
        ("value", 10**400),
        ("unit", "V"),
        ("kind", "image"),
        ("observed_at", "2026-10-25T12:00:00"),
        ("observed_at", "2026-10-25T12:00:00+03:00"),
        ("event_id", "x" * 129),
    ],
)
def test_reject_invalid(field, value):
    data = valid()
    data[field] = value
    with pytest.raises(ValueError):
        Observation.parse(data)


def test_no_raw_media_or_unknown_fields():
    data = valid()
    data["raw_frame"] = "never store source media"
    with pytest.raises(ValueError):
        Observation.parse(data)
