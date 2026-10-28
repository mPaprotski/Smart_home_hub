"""Loopback-only HTTPS reverse-proxy/CA spike. No system trust or mDNS changes."""
import http.server
import json
import os
import socket
import ssl
import sys
import tempfile
import threading
import urllib.request
from contextlib import contextmanager
from datetime import datetime,timedelta,timezone
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import emit


@contextmanager
def serving(handler,context=None):
    class QuietServer(http.server.ThreadingHTTPServer):
        daemon_threads=True
        def handle_error(self,request,client_address):pass
    server=QuietServer(('127.0.0.1',0),handler)
    if context:server.socket=context.wrap_socket(server.socket,server_side=True)
    thread=threading.Thread(target=server.serve_forever,kwargs={'poll_interval':.02},daemon=True)
    thread.start()
    try:yield server.server_port
    finally:server.shutdown();server.server_close();thread.join(timeout=2)


class Backend(http.server.BaseHTTPRequestHandler):
    def log_message(self,*args):pass
    def do_GET(self):
        body=b'{"marker":"smarthome-spike-backend","production_core":false}'
        self.send_response(200);self.send_header('Content-Length',str(len(body)));self.end_headers();self.wfile.write(body)


def request(port,context,hostname='smarthome.local',path='/health',method='GET'):
    with socket.create_connection(('127.0.0.1',port),timeout=2) as raw:
        with context.wrap_socket(raw,server_hostname=hostname) as sock:
            sock.sendall(f'{method} {path} HTTP/1.0\r\nHost: {hostname}\r\nConnection: close\r\n\r\n'.encode())
            data=bytearray()
            while len(data)<16384:
                chunk=sock.recv(4096)
                if not chunk:break
                data.extend(chunk)
            return bytes(data)


def probe():
    try:
        from cryptography import x509
        from cryptography.hazmat.primitives import hashes,serialization
        from cryptography.hazmat.primitives.asymmetric import rsa
        from cryptography.x509.oid import NameOID,ExtendedKeyUsageOID
    except ImportError:
        return emit('D09-HTTPS','BLOCKED',[],reason='cryptography package missing')
    now=datetime.now(timezone.utc)
    ca_key=rsa.generate_private_key(public_exponent=65537,key_size=2048)
    ca_name=x509.Name([x509.NameAttribute(NameOID.COMMON_NAME,'Ephemeral SmartHome Probe CA')])
    ca=(x509.CertificateBuilder().subject_name(ca_name).issuer_name(ca_name).public_key(ca_key.public_key())
        .serial_number(x509.random_serial_number()).not_valid_before(now-timedelta(minutes=1))
        .not_valid_after(now+timedelta(days=1)).add_extension(x509.BasicConstraints(ca=True,path_length=0),critical=True)
        .add_extension(x509.KeyUsage(digital_signature=True,key_encipherment=False,content_commitment=False,
            data_encipherment=False,key_agreement=False,key_cert_sign=True,crl_sign=True,encipher_only=False,decipher_only=False),critical=True)
        .sign(ca_key,hashes.SHA256()))
    ca_pem=ca.public_bytes(serialization.Encoding.PEM).decode()
    trusted=ssl.create_default_context(cadata=ca_pem)
    untrusted=ssl.create_default_context()
    serials=[]
    try:
        with tempfile.TemporaryDirectory(prefix='smarthome-tls-') as directory:
            def tls_context(index):
                key=rsa.generate_private_key(public_exponent=65537,key_size=2048)
                cert=(x509.CertificateBuilder().subject_name(x509.Name([x509.NameAttribute(NameOID.COMMON_NAME,'smarthome.local')]))
                      .issuer_name(ca_name).public_key(key.public_key()).serial_number(x509.random_serial_number())
                      .not_valid_before(now-timedelta(minutes=1)).not_valid_after(now+timedelta(hours=1))
                      .add_extension(x509.BasicConstraints(ca=False,path_length=None),critical=True)
                      .add_extension(x509.SubjectAlternativeName([x509.DNSName('smarthome.local')]),critical=False)
                      .add_extension(x509.ExtendedKeyUsage([ExtendedKeyUsageOID.SERVER_AUTH]),critical=False)
                      .sign(ca_key,hashes.SHA256()))
                serials.append(cert.serial_number)
                key_path=Path(directory)/f'leaf-{index}.key';cert_path=Path(directory)/f'leaf-{index}.crt'
                key_bytes=key.private_bytes(serialization.Encoding.PEM,serialization.PrivateFormat.PKCS8,serialization.NoEncryption())
                fd=os.open(key_path,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
                with os.fdopen(fd,'wb') as f:f.write(key_bytes)
                cert_path.write_bytes(cert.public_bytes(serialization.Encoding.PEM))
                assert key_path.stat().st_mode&0o777==0o600
                context=ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER);context.minimum_version=ssl.TLSVersion.TLSv1_2
                context.load_cert_chain(cert_path,key_path)
                return context
            with serving(Backend) as backend_port:
                class Proxy(http.server.BaseHTTPRequestHandler):
                    def log_message(self,*args):pass
                    def do_GET(self):
                        if self.path!='/health':
                            self.send_error(404);return
                        # No environment proxy, no arbitrary upstream/SSRF, no raw client input forwarding.
                        opener=urllib.request.build_opener(urllib.request.ProxyHandler({}))
                        with opener.open(f'http://127.0.0.1:{backend_port}/health',timeout=2) as upstream:
                            body=upstream.read(4096)
                        self.send_response(200);self.send_header('Content-Type','application/json')
                        self.send_header('Content-Length',str(len(body)));self.end_headers();self.wfile.write(body)
                    def do_POST(self):self.send_error(405)
                with serving(Proxy,tls_context(1)) as port:
                    good=request(port,trusted)
                    assert b'200 OK' in good and b'smarthome-spike-backend' in good
                    for context,hostname in [(untrusted,'smarthome.local'),(trusted,'wrong.local')]:
                        try:request(port,context,hostname)
                        except ssl.SSLCertVerificationError:pass
                        else:raise AssertionError('TLS verification bypassed')
                    assert b'404 Not Found' in request(port,trusted,path='/api/v1/lights')
                    assert b'405 Method Not Allowed' in request(port,trusted,method='POST')
                with serving(Proxy,tls_context(2)) as port:
                    assert b'200 OK' in request(port,trusted)
                assert serials[0]!=serials[1]
        emit('D09-HTTPS','PASS-HOST',['ephemeral CA→server SAN smarthome.local',
             'loopback HTTPS proxy→loopback backend','untrusted CA rejected','wrong hostname rejected',
             'control routes absent/mutations rejected','certificate replacement under same CA',
             'private key mode0600/temp cleanup'],bind_scope='127.0.0.1 only',
             mdns='BLOCKED_NO_PI_LAN_TARGET',phone_trust='BLOCKED_NO_CLIENT',
             production_proxy='NOT_RUN_NO_CADDY_NGINX',system_trust_changed=False,
             ca_private_key_persisted=False)
    except PermissionError:
        emit('D09-HTTPS','BLOCKED',[],reason='sandbox denies loopback listen',system_trust_changed=False)


if __name__=='__main__':probe()
