"""Disposable Compose HTTPS verification under the Pi core memory limit."""
import contextlib
import json
import os
from pathlib import Path
import subprocess
import tempfile

from auth_smoke import PASSWORD, ROOT, certificates, free_port, scenario

PROJECT = 'smarthome-d21-smoke'
BASE = ['docker', 'compose', '-p', PROJECT, '-f', 'deploy/docker/compose.dev.yaml',
        '-f', 'deploy/docker/compose.sqlite.yaml']
TLS = [*BASE, '-f', 'deploy/docker/compose.auth.yaml']


def main():
    port = free_port()
    origin = 'https://127.0.0.1:' + str(port)
    env = dict(os.environ, HUB_DEV_PORT=str(port), HUB_PUBLIC_ORIGIN=origin)

    def run(args, data=None):
        proc = subprocess.run(args, input=data, capture_output=True, cwd=ROOT, env=env, timeout=100)
        assert proc.returncode == 0, proc.stderr.decode()
        return proc.stdout.decode().strip()

    assert not run([*BASE, 'ps', '--all', '--quiet']), 'test project already exists'
    exists = subprocess.run(['docker', 'volume', 'inspect', PROJECT + '_hub-data'], capture_output=True)
    assert exists.returncode != 0, 'refusing existing volume'
    peaks = []
    try:
        with tempfile.TemporaryDirectory(prefix='hub-compose-tls-') as directory:
            cert = certificates(Path(directory) / 'tls')
            run([*BASE, 'up', '-d', '--no-build', '--wait', '--wait-timeout', '45'])
            run([*BASE, 'exec', '-T', 'hub', '/app/hub', 'admin', 'bootstrap', '--config',
                 '/app/config.sqlite.yaml', '--username', 'admin', '--password-stdin'], (PASSWORD + '\n').encode())
            for name in ('server.pem', 'server.key', 'ca.pem'):
                run([*BASE, 'exec', '-T', 'hub', 'sh', '-ec',
                     'umask 077; mkdir -p /data/hub/tls; cat > /data/hub/tls/' + name],
                    (cert / name).read_bytes())

            @contextlib.contextmanager
            def process():
                run([*TLS, 'up', '-d', '--no-build', '--wait', '--wait-timeout', '45'])
                container = run([*TLS, 'ps', '--quiet', 'hub'])
                info = json.loads(run(['docker', 'inspect', container]))[0]
                assert info['HostConfig']['Memory'] == 128 * 1024 * 1024
                assert info['HostConfig']['MemorySwap'] == info['HostConfig']['Memory']
                assert info['HostConfig']['ReadonlyRootfs']
                assert info['Config']['User'] == '10001:10001'
                assert 'GOMEMLIMIT=96MiB' in info['Config']['Env']
                assert run([*TLS, 'port', 'hub', '8080']) == '127.0.0.1:' + str(port)
                yield
                peaks.append(int(run([*TLS, 'exec', '-T', 'hub', 'cat', '/sys/fs/cgroup/memory.peak'])))
                assert peaks[-1] < 128 * 1024 * 1024
                run([*TLS, 'stop', '--timeout', '10', 'hub'])
                info = json.loads(run(['docker', 'inspect', container]))[0]
                assert not info['State']['OOMKilled'] and info['State']['ExitCode'] == 0

            result = scenario(process, process, origin, cert / 'ca.pem')
            result.update(memory_limit_bytes=128 * 1024 * 1024, memory_swap_equal_limit=True,
                          peak_memory_bytes=peaks, readonly_rootfs=True, uid=10001, ca_verified=True,
                          oom_killed=False, simulated=True, hardware_validated=False)
    finally:
        run([*TLS, 'down', '--volumes', '--remove-orphans', '--timeout', '10'])
    result['test_volume_removed'] = True
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
