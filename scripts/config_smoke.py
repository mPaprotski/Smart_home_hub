"""Black-box configuration failure/redaction checks; never opens physical devices."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
SENTINEL = "synthetic-never-log-token"


def main():
    env = {k: v for k, v in os.environ.items() if not k.startswith("HUB_")}
    env.update(SESSION_SECRET=SENTINEL, ML_SERVICE_TOKEN=SENTINEL, MQTT_PASSWORD=SENTINEL)
    cases = {"unknown_key": "unknown: " + SENTINEL,
             "duplicate_key": "profile: test\nprofile: test",
             "production_closed": "profile: production",
             "hardware_closed": "adapter_mode: hardware",
             "public_bind_closed": "server:\n  host: 0.0.0.0",
             "nan_brightness": "lighting:\n  zones:\n    - id: test\n      count: 1\n      brightness: .nan"}
    output = {}
    with tempfile.TemporaryDirectory(prefix="hub-config-") as directory:
        path = Path(directory) / "test.yaml"
        for name, data in cases.items():
            path.write_text(data, encoding="utf-8")
            run = subprocess.run([str(ROOT / "bin/hub"), "--config", str(path)], env=env,
                                 capture_output=True, text=True, timeout=5)
            assert run.returncode == 2, name
            assert SENTINEL not in run.stdout + run.stderr, name
            entries = [json.loads(line) for line in run.stderr.splitlines()]
            assert all(e["msg"] != "http_started" for e in entries), name
            output[name] = dict(exit_code=2, listener_not_started=True,
                                safe_error=entries[-1]["error"])
    run = subprocess.run([str(ROOT / "bin/hub"), "--config", str(ROOT / "config.example.yaml"),
                          "--check-config"], env=env, capture_output=True, text=True, timeout=5)
    assert run.returncode == 0
    assert SENTINEL not in run.stderr + run.stdout
    output["valid_redacted"] = dict(exit_code=0, secret_absent=True)
    print(json.dumps(output, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
