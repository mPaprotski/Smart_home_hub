"""Actual compiled Hub process smoke, isolated loopback listener and synthetic data only."""
import json
import os
from pathlib import Path
import signal
import subprocess
import tempfile
import time
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[1]


def scenario(storage):
    # Pin all infrastructure overrides so a developer shell cannot change the test's scope.
    env = {k: v for k, v in os.environ.items() if not k.startswith("HUB_")}
    env.update(HUB_PROFILE="test", HUB_ADAPTER_MODE="simulation", HUB_STORAGE=storage,
               HUB_SERVER_HOST="127.0.0.1", HUB_SERVER_PORT="0", SESSION_SECRET="synthetic-secret-sentinel")
    with tempfile.TemporaryFile(mode="w+") as log:
        proc = subprocess.Popen([str(ROOT / "bin/hub"), "--config", str(ROOT / "config.example.yaml")],
                                cwd=ROOT, env=env, stderr=log, stdout=subprocess.DEVNULL)
        address = None
        result = {}
        try:
            deadline = time.monotonic() + 10
            while time.monotonic() < deadline:
                log.seek(0)
                entries = [json.loads(line) for line in log if line.strip()]
                address = next((e["address"] for e in entries if e["msg"] == "http_started"), None)
                if address:
                    break
                if proc.poll() is not None:
                    raise AssertionError("hub exited during startup")
                time.sleep(.02)
            assert address, "startup deadline exceeded"
            base = "http://" + address
            opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
            for path in ("/health", "/ready", "/api/v1/version"):
                try:
                    response = opener.open(base + path, timeout=2)
                except urllib.error.HTTPError as exc:
                    response = exc
                with response:
                    result[path] = dict(status=response.status, body=json.load(response))
            assert result["/ready"]["body"]["ready"] == (storage == "memory_fake")
            assert result["/ready"]["status"] == (200 if storage == "memory_fake" else 503)
            assert not result["/ready"]["body"]["persistent"]
            assert result["/ready"]["body"]["simulated"]
            # Observe a real counter, rather than assuming a running ticker produced events.
            deadline = time.monotonic() + 5
            published = 0
            while time.monotonic() < deadline:
                with opener.open(base + "/metrics", timeout=2) as response:
                    metrics = response.read().decode()
                published = int(next(line.split()[1] for line in metrics.splitlines()
                                     if line.startswith("hub_events_published_total ")))
                if published >= 5:
                    break
                time.sleep(.02)
            assert published >= 5, "no synthetic events published"
            try:
                opener.open(urllib.request.Request(base + "/api/v1/commands", method="POST"), timeout=2)
                raise AssertionError("control route unexpectedly exists")
            except urllib.error.HTTPError as exc:
                assert exc.code == 404
            proc.send_signal(signal.SIGTERM)
            assert proc.wait(timeout=7) == 0, "unclean shutdown"
            try:
                opener.open(base + "/health", timeout=.5)
                raise AssertionError("listener survived shutdown")
            except urllib.error.URLError:
                pass
            log.seek(0)
            logs = log.read()
            assert "synthetic-secret-sentinel" not in logs
            stopped = [json.loads(line) for line in logs.splitlines() if '"hub_stopped"' in line]
            assert stopped and stopped[-1]["observations"] >= 5, "events never reached Core observer"
            result.update(published=published, observed=stopped[-1]["observations"], exit_code=0,
                          listener_closed=True, secret_redacted=True)
            return result
        finally:
            if proc.poll() is None:
                proc.kill()
                proc.wait(timeout=3)


if __name__ == "__main__":
    result = {kind: scenario(kind) for kind in ("memory_fake", "not_implemented")}
    print(json.dumps(result, ensure_ascii=False, indent=2))
