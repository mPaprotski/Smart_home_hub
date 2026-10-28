"""Read-only redacted environment inventory. No host names, IPs or secret env."""
import glob
import importlib.util
import os
import platform
import shutil
import subprocess
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from common import emit


def inventory():
    os_info={}
    for line in Path('/etc/os-release').read_text().splitlines():
        if line.startswith(('ID=','VERSION_ID=','NAME=','VERSION=')):
            key,value=line.split('=',1);os_info[key]=value.strip('"')
    model=None
    for p in ['/proc/device-tree/model','/sys/firmware/devicetree/base/model']:
        if Path(p).exists(): model=Path(p).read_bytes().rstrip(b'\0').decode(errors='replace');break
    versions={}
    commands={'python':[sys.executable,'--version'],'pipewire':['pipewire','--version'],
              'wireplumber':['wireplumber','--version'],'ffmpeg':['ffmpeg','-version'],
              'openssl':['openssl','version'],'go':['/snap/go/current/bin/go','version']}
    for key,cmd in commands.items():
        if not Path(cmd[0]).exists() and not shutil.which(cmd[0]): continue
        try:
            r=subprocess.run(cmd,capture_output=True,text=True,timeout=5)
            versions[key]={'exit_code':r.returncode,'version':r.stdout.splitlines()[:2]}
        except (OSError,subprocess.TimeoutExpired): versions[key]={'available':False}
    return {'os':os_info,'architecture':platform.machine(),'pi_model':model,
        'is_pi4':bool(model and 'Raspberry Pi 4' in model),
        'device_counts':{k:len(glob.glob(v)) for k,v in {'audio':'/dev/snd/*','camera':'/dev/video*','gpio':'/dev/gpiochip*','i2c':'/dev/i2c-*','spi':'/dev/spidev*'}.items()},
        'systemd_pid1':Path('/proc/1/comm').read_text().strip()=='systemd',
        'versions':versions,'display_environment_present':{k:bool(os.getenv(k)) for k in ['DISPLAY','WAYLAND_DISPLAY']},
        'modules':{m:importlib.util.find_spec(m) is not None for m in ['PIL','onnxruntime','numpy']}}


if __name__=='__main__':
    data=inventory()
    emit('D06-environment','PASS-HOST',['read-only OS/architecture/version/device inventory'],
         environment=data,pi_installation='NOT_RUN',
         target_status='PI_DETECTED_REQUIRES_REVIEW' if data['is_pi4'] else 'BLOCKED_NO_PI')
