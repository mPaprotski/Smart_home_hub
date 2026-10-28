#!/usr/bin/env python3
"""Run bounded feasibility probes. PASS-HOST/SIM never means PASS-PI or M0 ready."""
import argparse
import json
import os
import shutil
import subprocess
import sys
import time
from pathlib import Path
from common import ROOT


def go_binary():
    direct=Path('/snap/go/current/bin/go')
    return str(direct) if direct.exists() else shutil.which('go')


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output',type=Path,default=ROOT/'experiments/results/latest.json')
    parser.add_argument('--require-m0',action='store_true')
    parser.add_argument('--live-capture',action='store_true')
    args=parser.parse_args()
    cache=ROOT/'experiments/.cache/go-build';cache.mkdir(parents=True,exist_ok=True)
    env=dict(os.environ,GOCACHE=str(cache),GO111MODULE='off',GOTOOLCHAIN='local')
    probes=[('environment',[sys.executable,'experiments/day006/environment_probe.py'],20),
      ('pcm',[sys.executable,'experiments/day004/pcm_probe.py'],15),
      ('pipewire',[sys.executable,'experiments/day004/pipewire_probe.py'],20),
      ('screen',[sys.executable,'experiments/day005/screen_probe.py']+(['--live-capture'] if args.live_capture else []),10),
      ('inference',[sys.executable,'experiments/day005/inference_probe.py'],10),
      ('thermal',[sys.executable,'experiments/day007/thermal_policy_probe.py'],5),
      ('hil',[sys.executable,'experiments/day008/hil_preflight.py'],5),
      ('https',[sys.executable,'experiments/day009/https_probe.py'],15)]
    go=go_binary()
    if go:probes.insert(1,('zones',[go,'run','experiments/day004/zone_frame_probe.go'],90))
    results={};started=time.monotonic()
    for name,cmd,timeout in probes:
        begin=time.monotonic()
        try:
            r=subprocess.run(cmd,cwd=ROOT,env=env,capture_output=True,text=True,timeout=timeout)
            try:result=json.loads(r.stdout)
            except json.JSONDecodeError:result={'status':'FAIL','reason':'probe produced no valid JSON','exit_code':r.returncode}
            if r.returncode!=0:result.update(status='FAIL',exit_code=r.returncode)
        except subprocess.TimeoutExpired:result={'status':'FAIL','reason':'bounded timeout'}
        results[name]=dict(result,runner_elapsed_ms=round((time.monotonic()-begin)*1000,3))
        print(f'{name}: {results[name]["status"]}',flush=True)
    if not go:results['zones']={'status':'BLOCKED','reason':'Go toolchain unavailable'}
    unit_commands=[('python_unittests',[sys.executable,'-m','unittest','discover','-s','experiments/tests','-v'])]
    if go:unit_commands.append(('go_tests',[go,'test','-race','./experiments/day004']))
    for name,cmd in unit_commands:
        try:
            r=subprocess.run(cmd,cwd=ROOT,env=env,capture_output=True,text=True,timeout=90)
            results[name]={'status':'PASS-HOST' if r.returncode==0 else 'FAIL','exit_code':r.returncode,
                           'summary':(r.stdout+r.stderr).replace(str(ROOT),'<workspace>')[-10000:]}
        except subprocess.TimeoutExpired:results[name]={'status':'FAIL','reason':'test timeout'}
        print(f'{name}: {results[name]["status"]}',flush=True)
    result={'client_date':'2026-10-25','environment_kind':'developer_host_not_pi',
       'suite':'days004-010-feasibility','results':results,
       'elapsed_s':round(time.monotonic()-started,3),
       'm0':{'status':'M0_NOT_READY','reason':'Pi installation/hardware/LAN/mobile evidence unavailable; host/simulation checks do not close these gates'}}
    args.output.parent.mkdir(parents=True,exist_ok=True)
    args.output.write_text(json.dumps(result,ensure_ascii=False,indent=2,allow_nan=False)+'\n')
    if any(r['status']=='FAIL' for r in results.values()):return 1
    return 2 if args.require_m0 else 0


if __name__=='__main__':raise SystemExit(main())
