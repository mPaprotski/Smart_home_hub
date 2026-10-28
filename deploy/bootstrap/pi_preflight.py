#!/usr/bin/env python3
"""Read-only Pi target gate. Does not flash disks, reboot, install or toggle GPIO."""
import argparse
import glob
import json
import platform
from pathlib import Path


def inspect(root=Path('/')):
    model=None
    for relative in ['proc/device-tree/model','sys/firmware/devicetree/base/model']:
        p=root/relative
        if p.exists(): model=p.read_bytes().rstrip(b'\0').decode(errors='replace');break
    temp=root/'sys/class/thermal/thermal_zone0/temp'
    temperature=None
    if temp.exists():
        try: temperature=int(temp.read_text().strip())/1000
        except (ValueError,OSError): pass
    memory_total_kib=None
    meminfo=root/'proc/meminfo'
    if meminfo.exists():
        for line in meminfo.read_text().splitlines():
            if line.startswith('MemTotal:'):
                try: memory_total_kib=int(line.split()[1])
                except (ValueError,IndexError): pass
                break
    is_pi4=bool(model and 'Raspberry Pi 4' in model)
    matches_2gb=bool(is_pi4 and memory_total_kib is not None and 1536*1024<=memory_total_kib<=2100*1024)
    return {'target':'Raspberry Pi 4 Model B, only 2GB','target_ram_gb':2,'memory_total_kib':memory_total_kib,'target_ram_matches':matches_2gb,'is_pi4':is_pi4,'architecture':platform.machine(),
        'temperature_c':temperature if is_pi4 else None,
        'devices':{kind:len(list((root/'dev').glob(pattern))) for kind,pattern in
                   [('gpio','gpiochip*'),('i2c','i2c-*'),('spi','spidev*'),('camera','video*')]},
        'status':'TARGET_DETECTED_NOT_ACCEPTED' if matches_2gb else ('BLOCKED_RAM_UNVERIFIED_OR_MISMATCH' if is_pi4 else 'BLOCKED_NO_PI'),
        'mutations_performed':False,'secrets_collected':False}


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--require-pi',action='store_true')
    args=parser.parse_args();result=inspect();print(json.dumps(result,indent=2))
    raise SystemExit(2 if args.require_pi and not result['target_ram_matches'] else 0)
