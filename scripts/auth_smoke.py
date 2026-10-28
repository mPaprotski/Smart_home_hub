"""Compiled HTTPS/session/token/audit smoke; all credentials and TLS keys are disposable."""
import contextlib
import http.cookiejar
import json
import os
from pathlib import Path
import signal
import socket
import ssl
import subprocess
import tempfile
import time
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
PASSWORD = "Synthetic only HTTPS password sentinel 42"


def free_port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def certificates(directory):
    directory = Path(directory)
    directory.mkdir(mode=0o700, parents=True, exist_ok=True)
    def run(*args):
        subprocess.run(["openssl", *args], cwd=directory, check=True,
                       capture_output=True, text=True, timeout=20)
    run("req", "-x509", "-newkey", "rsa:2048", "-nodes", "-keyout", "ca.key", "-out", "ca.pem",
        "-days", "1", "-subj", "/CN=Disposable Hub Test CA", "-addext", "basicConstraints=critical,CA:TRUE")
    run("req", "-newkey", "rsa:2048", "-nodes", "-keyout", "server.key", "-out", "server.csr",
        "-subj", "/CN=localhost")
    (directory / "extensions").write_text("subjectAltName=IP:127.0.0.1,DNS:localhost\nbasicConstraints=CA:FALSE\nextendedKeyUsage=serverAuth\n", encoding="utf-8")
    run("x509", "-req", "-in", "server.csr", "-CA", "ca.pem", "-CAkey", "ca.key", "-CAcreateserial",
        "-out", "server.pem", "-days", "1", "-extfile", "extensions")
    for name in ("ca.key", "server.key"):
        (directory / name).chmod(0o600)
    return directory


class Client:
    def __init__(self, base, ca):
        self.base = base
        self.jar = http.cookiejar.CookieJar()
        context = ssl.create_default_context(cafile=str(ca))
        self.opener = urllib.request.build_opener(urllib.request.ProxyHandler({}),
                                                  urllib.request.HTTPSHandler(context=context),
                                                  urllib.request.HTTPCookieProcessor(self.jar))

    def request(self, path, method="GET", body=None, csrf=None, bearer=None, origin=None):
        headers = {"X-Hub-Request": "1"}
        if method != "GET":
            headers["Origin"] = self.base if origin is None else origin
        if csrf:
            headers["X-CSRF-Token"] = csrf
        if bearer:
            headers["Authorization"] = "Bearer " + bearer
        data = None if body is None else json.dumps(body).encode()
        if data is not None:
            headers["Content-Type"] = "application/json"
        req = urllib.request.Request(self.base + path, data=data, headers=headers, method=method)
        try:
            response = self.opener.open(req, timeout=5)
        except urllib.error.HTTPError as exc:
            response = exc
        with response:
            raw = response.read()
            media = response.headers.get("Content-Type", "")
            value = json.loads(raw) if raw and "json" in media else raw.decode()
            return response.status, dict(response.headers), value


def scenario(first, restart, base, ca):
    browser = Client(base, ca)
    anonymous = Client(base, ca)
    result = {}
    with first():
        assert anonymous.request("/api/v1/auth/me")[0] == 401
        code, headers, login = browser.request("/api/v1/auth/login", "POST", {"username": "admin", "password": PASSWORD})
        assert code == 200, (code, login)
        csrf, user = login["csrf_token"], login["user_id"]
        for flag in ("Secure", "HttpOnly", "SameSite=Lax", "Path=/"):
            assert flag in headers["Set-Cookie"]
        session_secret = next(cookie.value for cookie in browser.jar if cookie.name == "__Host-shh_session")
        assert browser.request("/api/v1/auth/me")[0] == 200
        body = {"value": "Europe/Minsk", "expected_revision": 0, "command_id": "browser-setting"}
        assert browser.request("/api/v1/users/me/settings/timezone", "PUT", body, csrf="wrong")[0] == 403
        assert browser.request("/api/v1/users/me/settings/timezone", "PUT", body, csrf=csrf, origin="https://evil.example")[0] == 403
        assert browser.request("/api/v1/users/me/settings/timezone", "PUT", body, csrf=csrf)[0] == 200
        code, _, created = browser.request("/api/v1/auth/tokens", "POST", {"name": "writer", "scopes": ["settings:write"], "bindings": ["user:" + user]}, csrf=csrf)
        assert code == 201, created
        token, token_id = created["token"], created["metadata"]["id"]
        code, _, listed = browser.request("/api/v1/auth/tokens")
        assert code == 200 and token not in json.dumps(listed)
        assert browser.request("/api/v1/auth/me", bearer=token)[0] == 400
        assert anonymous.request("/api/v1/auth/me", bearer=token)[0] == 200
        result.update(login_https=True, csrf_origin_enforced=True, cookie_flags=True,
                      one_time_token_reveal=True, ambiguous_auth_rejected=True)
    with restart():
        assert browser.request("/api/v1/auth/me")[0] == 200, "session lost on restart"
        body = {"value": "UTC", "expected_revision": 1, "command_id": "machine-setting"}
        assert anonymous.request("/api/v1/users/me/settings/timezone", "PUT", body, bearer=token)[0] == 200
        assert browser.request("/api/v1/auth/tokens/revoke", "POST", {"id": token_id}, csrf=csrf)[0] == 204
        assert anonymous.request("/api/v1/users/me/settings/timezone", "PUT", body, bearer=token)[0] == 401
        code, _, audit = browser.request("/api/v1/auth/audit")
        assert code == 200 and len(audit) >= 5
        text = json.dumps(audit)
        for secret in (PASSWORD, token, session_secret, csrf):
            assert secret not in text
        assert browser.request("/api/v1/auth/logout", "POST", csrf=csrf)[0] == 204
        assert browser.request("/api/v1/auth/me")[0] == 401
        code, _, ready = anonymous.request("/ready")
        assert code == 200 and ready["persistent"] and ready["simulated"]
        assert any(d["name"] == "auth" and d["status"] == "READY" for d in ready["dependencies"])
        result.update(session_after_restart=True, token_after_restart=True, setting_revision=2,
                      revoke_before_cached_retry=True, audit_rows=len(audit), audit_redacted=True,
                      logout=True, auth_required_ready=200)
    return result


def main():
    with tempfile.TemporaryDirectory(prefix="hub-auth-") as directory:
        private = Path(directory)
        cert = certificates(private / "tls")
        port = free_port()
        base = "https://127.0.0.1:" + str(port)
        env = {k: v for k, v in os.environ.items() if not k.startswith("HUB_")}
        env.update(HUB_DB_PATH=str(private / "data" / "hub.db"), HUB_SERVER_HOST="127.0.0.1",
                   HUB_SERVER_PORT=str(port), HUB_TLS_CERT=str(cert / "server.pem"),
                   HUB_TLS_KEY=str(cert / "server.key"), HUB_PUBLIC_ORIGIN=base)
        config = ROOT / "configs/auth.dev.yaml"
        created = subprocess.run([str(ROOT / "bin/hub"), "admin", "bootstrap", "--config", str(config),
                                  "--username", "admin", "--password-stdin"], input=PASSWORD + "\n",
                                 capture_output=True, text=True, env=env, timeout=20)
        assert created.returncode == 0, created.stderr
        counts = []

        @contextlib.contextmanager
        def process():
            with tempfile.TemporaryFile(mode="w+") as log:
                proc = subprocess.Popen([str(ROOT / "bin/hub"), "--config", str(config)], env=env,
                                        cwd=ROOT, stdout=subprocess.DEVNULL, stderr=log)
                try:
                    deadline = time.monotonic() + 15
                    client = Client(base, cert / "ca.pem")
                    while True:
                        try:
                            if client.request("/ready")[0] == 200:
                                break
                        except OSError:
                            pass
                        assert proc.poll() is None, "Hub exited during startup"
                        assert time.monotonic() < deadline, "startup deadline"
                        time.sleep(.03)
                    subprocess.run([str(ROOT / "bin/hub"), "--healthcheck", base + "/ready",
                                    "--healthcheck-ca", str(cert / "ca.pem")], check=True, env=env, timeout=5)
                    yield
                    deadline = time.monotonic() + 3
                    while True:
                        _, _, metrics = client.request("/metrics")
                        count = int(next(line.split()[1] for line in metrics.splitlines() if line.startswith("hub_events_published_total ")))
                        if count >= 5:
                            break
                        assert time.monotonic() < deadline
                        time.sleep(.02)
                    proc.send_signal(signal.SIGTERM)
                    assert proc.wait(timeout=7) == 0
                    log.seek(0)
                    text = log.read()
                    assert PASSWORD not in text and "$argon2" not in text
                    stopped = [json.loads(line) for line in text.splitlines() if '"hub_stopped"' in line]
                    assert stopped and stopped[-1]["observations"] >= 5
                    counts.append(stopped[-1]["observations"])
                finally:
                    if proc.poll() is None:
                        proc.kill()
                        proc.wait(timeout=3)
        result = scenario(process, process, base, cert / "ca.pem")
        result.update(process_exit=0, observed_each_process=counts, ca_verified=True, tls_keys_removed=True)
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
