"""Isolated dev Compose smoke. Requires a prebuilt image for project smarthome-d15."""
import json
import os
from pathlib import Path
import subprocess
import time
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
COMMAND = ["docker", "compose", "-p", "smarthome-d15", "-f", "deploy/docker/compose.dev.yaml"]
ENV = dict(os.environ, HUB_DEV_PORT="0")


def compose(*args):
    return subprocess.check_output([*COMMAND, *args], cwd=ROOT, env=ENV, text=True,
                                   stderr=subprocess.STDOUT, timeout=90).strip()


def inspect(container):
    return json.loads(subprocess.check_output(["docker", "inspect", container], text=True))[0]


def main():
    # Refuse to stop/replace a container from an existing developer session.
    assert not compose("ps", "--all", "--quiet"), "test project already has containers"
    result = {}
    try:
        compose("up", "--detach", "--no-build", "--wait", "--wait-timeout", "45")
        container = compose("ps", "--quiet", "hub")
        details = inspect(container)
        assert details["Config"]["User"] == "10001:10001"
        assert details["HostConfig"]["ReadonlyRootfs"]
        assert details["State"]["Health"]["Status"] == "healthy"
        assert not details["HostConfig"]["Privileged"]
        address = compose("port", "hub", "8080")
        assert address.startswith("127.0.0.1:"), address
        opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
        base = "http://" + address

        def ready():
            with opener.open(base + "/ready", timeout=2) as response:
                data = json.load(response)
            assert data["ready"] and data["simulated"] and not data["persistent"]
            return data

        def await_ready():
            deadline = time.monotonic() + 10
            while True:
                try:
                    return ready()
                except (OSError, AssertionError):
                    if time.monotonic() >= deadline:
                        raise
                    time.sleep(.1)

        result["ready"] = await_ready()
        with opener.open(base + "/metrics", timeout=2) as response:
            metrics = response.read().decode()
        published = int(next(line.split()[1] for line in metrics.splitlines()
                             if line.startswith("hub_events_published_total ")))
        assert published >= 5
        first_started = details["State"]["StartedAt"]
        compose("restart", "--timeout", "7", "hub")
        assert inspect(container)["State"]["StartedAt"] != first_started
        restarted_address = compose("port", "hub", "8080")
        assert restarted_address.startswith("127.0.0.1:")
        base = "http://" + restarted_address
        await_ready()
        # Ensure a complete batch reaches the observer in the restarted process too.
        deadline = time.monotonic() + 5
        while time.monotonic() < deadline:
            with opener.open(base + "/metrics", timeout=2) as response:
                text = response.read().decode()
            if any(line.startswith("hub_events_published_total ") and int(line.split()[1]) >= 5
                   for line in text.splitlines()):
                break
            time.sleep(.05)
        else:
            raise AssertionError("restarted process did not publish")
        compose("stop", "--timeout", "7", "hub")
        details = inspect(container)
        assert details["State"]["ExitCode"] == 0
        logs = compose("logs", "--no-color", "--no-log-prefix", "hub")
        entries = [json.loads(line) for line in logs.splitlines() if line.startswith("{")]
        stopped = [e for e in entries if e.get("msg") == "hub_stopped"]
        assert len(stopped) == 2 and all(e["observations"] >= 5 for e in stopped)
        result.update(published_before_restart=published, restart_passed=True,
                      nonroot=True, read_only=True, loopback_publish=True,
                      shutdown_exit=0, observed_each_process=[e["observations"] for e in stopped])
    finally:
        compose("down", "--remove-orphans", "--timeout", "7")
    assert not compose("ps", "--all", "--quiet"), "containers remained after down"
    result["cleanup_passed"] = True
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
