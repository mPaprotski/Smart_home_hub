"""Native day16-18 black-box smoke. Uses a private temporary DB and synthetic password."""
import concurrent.futures
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
PASSWORD = "Synthetic only password sentinel 42"
OPENER = urllib.request.build_opener(urllib.request.ProxyHandler({}))


def main():
    env = {k: v for k, v in os.environ.items() if not k.startswith("HUB_")}
    env.update(HUB_SERVER_HOST="127.0.0.1", HUB_SERVER_PORT="0")
    result = {}
    with tempfile.TemporaryDirectory(prefix="hub-persistent-") as directory:
        private = Path(directory)
        env["HUB_DB_PATH"] = str(private / "data" / "hub.db")

        def admin(action, password=PASSWORD):
            return subprocess.run([str(ROOT / "bin/hub"), "admin", action, "--config",
                                   str(ROOT / "configs/sqlite.dev.yaml"), "--username", "admin",
                                   "--password-stdin"], input=password + "\n", capture_output=True,
                                  text=True, env=env, timeout=20)

        created = admin("bootstrap")
        assert created.returncode == 0, created.stderr
        duplicate = admin("bootstrap")
        assert duplicate.returncode != 0 and "ALREADY_INITIALIZED" in duplicate.stderr
        verified = admin("verify")
        assert verified.returncode == 0, verified.stderr
        assert admin("verify", "wrong").returncode != 0
        assert PASSWORD not in created.stdout + created.stderr + duplicate.stdout + duplicate.stderr
        result["bootstrap_once"] = True
        result["cli_verification"] = True

        def process_scenario(required=False):
            config = ROOT / "configs/sqlite.dev.yaml"
            if required:
                config = private / "required.yaml"
                config.write_text("storage: sqlite\nauth:\n  required: true\n", encoding="utf-8")
            with tempfile.TemporaryFile(mode="w+") as log:
                proc = subprocess.Popen([str(ROOT / "bin/hub"), "--config", str(config)],
                                        cwd=ROOT, env=env, stdout=subprocess.DEVNULL, stderr=log)
                try:
                    address = None
                    deadline = time.monotonic() + 15
                    while time.monotonic() < deadline:
                        log.seek(0)
                        entries = [json.loads(line) for line in log if line.strip()]
                        address = next((e["address"] for e in entries if e["msg"] == "http_started"), None)
                        if address:
                            break
                        assert proc.poll() is None, "startup failed"
                        time.sleep(.02)
                    assert address
                    base = "http://" + address

                    def request(path, body=None, headers=None):
                        data = None if body is None else json.dumps(body).encode()
                        req = urllib.request.Request(base + path, data=data, headers=headers or {})
                        try:
                            response = OPENER.open(req, timeout=5)
                        except urllib.error.HTTPError as exc:
                            response = exc
                        with response:
                            raw = response.read().decode()
                            return response.status, dict(response.headers), raw

                    code, _, raw = request("/ready")
                    ready = json.loads(raw)
                    assert code == (503 if required else 200)
                    assert ready["persistent"] and ready["simulated"]
                    code, _, raw = request("/api/v1/version")
                    version = json.loads(raw)
                    assert code == 200 and version["persistent"] and version["auth"] == "CREDENTIALS_ONLY"
                    headers = {"Content-Type": "application/json", "X-Forwarded-For": "198.51.100.1"}
                    code, response_headers, raw = request("/api/v1/auth/login",
                                                         {"username": "admin", "password": PASSWORD}, headers)
                    assert code == 503 and json.loads(raw)["code"] == "AUTH_SESSION_NOT_IMPLEMENTED"
                    assert "Set-Cookie" not in response_headers and PASSWORD not in raw
                    if required:
                        prior, _, _ = request("/api/v1/auth/login", {"username": "missing", "password": "wrong"}, headers)
                        assert prior == 429, "account cooldown did not survive process restart"
                    missing_name = "missing-required" if required else "missing"
                    invalid, _, raw = request("/api/v1/auth/login",
                                               {"username": missing_name, "password": "wrong"}, headers)
                    assert invalid == 401 and json.loads(raw)["code"] == "INVALID_CREDENTIALS"
                    limited = None
                    for _ in range(5):
                        limited, _, raw = request("/api/v1/auth/login",
                                                  {"username": missing_name, "password": "wrong"}, headers)
                    assert limited == 429
                    # Use unknown accounts and varying forwarded headers to exercise peer-IP/global bounds.
                    def flood(index):
                        return request("/api/v1/auth/login", {"username": "flood" + str(index), "password": "wrong"},
                                       {"Content-Type": "application/json", "X-Forwarded-For": "203.0.113." + str(index)})[0]
                    with concurrent.futures.ThreadPoolExecutor(max_workers=8) as pool:
                        codes = list(pool.map(flood, range(20)))
                        live, _, _ = request("/health")
                    assert live == 200 and 429 in codes
                    proc.send_signal(signal.SIGTERM)
                    assert proc.wait(timeout=7) == 0
                    log.seek(0)
                    logs = log.read()
                    assert PASSWORD not in logs and "$argon2" not in logs
                    stop = [json.loads(line) for line in logs.splitlines() if '"hub_stopped"' in line]
                    assert stop and stop[-1]["observations"] >= 5
                    return {"ready_status": 503 if required else 200, "persistent": True, "simulated": True,
                            "login_valid": 503, "login_invalid": invalid, "login_throttled": limited,
                            "flood_health": live, "flood_statuses": sorted(set(codes)),
                            "observed": stop[-1]["observations"], "shutdown_exit": 0, "no_session_or_secret": True, "persisted_cooldown_checked": required}
                finally:
                    if proc.poll() is None:
                        proc.kill()
                        proc.wait(timeout=3)

        result["first_process"] = process_scenario()
        assert admin("verify").returncode == 0, "admin lost after process shutdown"
        result["restart_process"] = process_scenario(required=True)
        assert admin("verify").returncode == 0, "admin lost after second process"
        # Private DB permissions and no plaintext password on disk.
        for file in (private / "data").iterdir():
            if file.is_file():
                assert file.stat().st_mode & 0o077 == 0
                assert PASSWORD.encode() not in file.read_bytes()
        result["private_data_and_no_plaintext"] = True
    result["temporary_data_cleaned"] = True
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
