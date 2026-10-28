"""Persistent dev Compose smoke with a test-owned volume; refuses existing resources."""
import concurrent.futures
import json
import os
from pathlib import Path
import subprocess
import time
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
PROJECT = "smarthome-d18-smoke"
COMMAND = ["docker", "compose", "-p", PROJECT, "-f", "deploy/docker/compose.dev.yaml",
           "-f", "deploy/docker/compose.sqlite.yaml"]
ENV = dict(os.environ, HUB_DEV_PORT="0")
PASSWORD = "Synthetic only password sentinel 42"


def compose(*args, password=None):
    return subprocess.run([*COMMAND, *args], cwd=ROOT, env=ENV, input=password, text=True,
                          capture_output=True, timeout=90, check=False)


def checked(*args, password=None):
    run = compose(*args, password=password)
    if run.returncode != 0:
        raise AssertionError("Compose command failed: " + " ".join(args) + "\n" + run.stderr)
    return run.stdout.strip()


def main():
    assert not checked("ps", "--all", "--quiet"), "test project already has containers"
    existing_volume = subprocess.run(["docker", "volume", "inspect", PROJECT + "_hub-data"],
                                     capture_output=True, text=True)
    assert existing_volume.returncode != 0, "refusing to alter an existing data volume"
    result = {}
    try:
        checked("up", "--detach", "--no-build", "--wait", "--wait-timeout", "45")
        args = ["exec", "-T", "hub", "/app/hub", "admin"]
        checked(*args, "bootstrap", "--config", "/app/config.sqlite.yaml", "--username", "admin",
                "--password-stdin", password=PASSWORD + "\n")
        duplicate = compose(*args, "bootstrap", "--config", "/app/config.sqlite.yaml", "--username", "admin",
                            "--password-stdin", password=PASSWORD + "\n")
        assert duplicate.returncode != 0 and "ALREADY_INITIALIZED" in duplicate.stderr

        def verify():
            checked(*args, "verify", "--config", "/app/config.sqlite.yaml", "--username", "admin",
                    "--password-stdin", password=PASSWORD + "\n")

        def ready():
            address = checked("port", "hub", "8080")
            assert address.startswith("127.0.0.1:")
            opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
            deadline = time.monotonic() + 15
            while True:
                try:
                    with opener.open("http://" + address + "/ready", timeout=2) as response:
                        data = json.load(response)
                    assert data["ready"] and data["persistent"] and data["simulated"]
                    return data
                except (OSError, AssertionError):
                    if time.monotonic() >= deadline:
                        raise
                    time.sleep(.1)

        ready()
        verify()
        address=checked("port","hub","8080")
        opener=urllib.request.build_opener(urllib.request.ProxyHandler({}))
        def flood(index):
            request=urllib.request.Request("http://"+address+"/api/v1/auth/login",data=json.dumps({"username":"container-flood"+str(index),"password":"wrong"}).encode(),headers={"Content-Type":"application/json"})
            try: response=opener.open(request,timeout=5)
            except urllib.error.HTTPError as exc: response=exc
            with response: response.read(); return response.status
        with concurrent.futures.ThreadPoolExecutor(max_workers=8) as pool:
            futures=[pool.submit(flood,i) for i in range(20)]
            with opener.open("http://"+address+"/health",timeout=2) as response:
                assert response.status==200
            flood_codes=[future.result() for future in futures]
        assert 429 in flood_codes
        result["flood_health_status"]=200
        result["flood_statuses"]=sorted(set(flood_codes))
        peak=compose("exec","-T","hub","cat","/sys/fs/cgroup/memory.peak")
        result["peak_memory_bytes"]=int(peak.stdout.strip()) if peak.returncode==0 and peak.stdout.strip().isdigit() else None
        if result["peak_memory_bytes"] is not None:
            assert result["peak_memory_bytes"] <=128*1024*1024
        container = checked("ps", "--quiet", "hub")
        details = json.loads(subprocess.check_output(["docker", "inspect", container], text=True))[0]
        assert details["Config"]["User"] == "10001:10001" and details["HostConfig"]["ReadonlyRootfs"]
        assert details["HostConfig"]["Memory"] == 128 * 1024 * 1024
        checked("stop", "--timeout", "7", "hub")
        stopped = json.loads(subprocess.check_output(["docker", "inspect", container], text=True))[0]
        assert stopped["State"]["ExitCode"] == 0 and not stopped["State"]["OOMKilled"]
        # Recreate container, retain the same named volume.
        checked("up", "--detach", "--no-build", "--force-recreate", "--wait", "--wait-timeout", "45", "hub")
        ready()
        verify()
        result.update(persistent_admin_after_recreate=True, ready=ready(), bootstrap_once=True,
                      nonroot=True, rootfs_read_only=True, memory_limit_mib=128, loopback_publish=True)
        # A wrong private data mount/path must fail, never fall back to memory.
        failure = compose("run", "--rm", "--no-deps", "-T", "-e", "HUB_DB_PATH=/app/unwritable/hub.db", "hub")
        assert failure.returncode != 0 and "http_started" not in failure.stderr
        result["unwritable_storage_closed"] = True
        # Compose run -v can lose the ro override while merging service volumes.
        # Use an isolated Engine container and inspect the effective mount.
        readonly_name=PROJECT+"-readonly"
        image=details["Config"]["Image"]
        subprocess.run(["docker","create","--name",readonly_name,
                        "--label","com.docker.compose.project="+PROJECT,
                        "--read-only","--network","none","--user","10001:10001",
                        "--cap-drop","ALL","--security-opt","no-new-privileges:true",
                        "--memory","128m","--memory-swap","128m",
                        "--volume",PROJECT+"_hub-data:/data:ro",
                        "--env","HUB_DB_PATH=/data/hub/hub.db",image,
                        "--config","/app/config.sqlite.yaml"],check=True,capture_output=True,text=True)
        try:
            mount=json.loads(subprocess.check_output(["docker","inspect",readonly_name],text=True))[0]["Mounts"][0]
            assert mount["Destination"]=="/data" and not mount["RW"]
            subprocess.run(["docker","start",readonly_name],check=True,capture_output=True)
            deadline=time.monotonic()+12
            while True:
                state=json.loads(subprocess.check_output(["docker","inspect",readonly_name],text=True))[0]["State"]
                if not state["Running"]:break
                assert time.monotonic()<deadline,"readonly container failed to reject storage"
                time.sleep(.1)
            logs=subprocess.check_output(["docker","logs",readonly_name],stderr=subprocess.STDOUT,text=True)
            assert state["ExitCode"]!=0 and "http_started" not in logs
            result["read_only_data_volume_closed"]=True
            result["read_only_data_mount_inspected"]=True
        finally:
            subprocess.run(["docker","rm","--force",readonly_name],check=True,capture_output=True)
        checked("stop", "--timeout", "7", "hub")
        logs = checked("logs", "--no-color", "--no-log-prefix", "hub")
        assert PASSWORD not in logs and "$argon2" not in logs
        assert '"hub_stopped"' in logs
    finally:
        checked("down", "--volumes", "--remove-orphans", "--timeout", "7")
    assert not checked("ps", "--all", "--quiet")
    result["test_volume_and_containers_cleaned"] = True
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
