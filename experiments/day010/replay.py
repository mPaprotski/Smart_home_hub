"""Repeat independent sandbox suites; compare outcomes, never relabel as Pi proof."""
import argparse
import json
import subprocess
import sys
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import ROOT,emit


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--live-capture',action='store_true');args=parser.parse_args()
    runs=[]
    for index in [1,2]:
        path=ROOT/f'experiments/results/run-{index}.json'
        cmd=[sys.executable,str(ROOT/'experiments/run_all.py'),'--output',str(path)]
        if args.live_capture:cmd.append('--live-capture')
        r=subprocess.run(cmd,cwd=ROOT,timeout=240)
        if r.returncode:raise SystemExit(r.returncode)
        runs.append(json.loads(path.read_text()))
    statuses=[{k:v['status'] for k,v in run['results'].items()} for run in runs]
    if statuses[0]!=statuses[1]:raise SystemExit('probe outcomes differ between replays')
    # Compare deterministic numeric/frame outputs, not variable timing or monitor startup silence.
    for key in ['synthetic_colors']:
        assert runs[0]['results']['screen'][key]==runs[1]['results']['screen'][key]
    assert runs[0]['results']['zones']['limited_current_a']==runs[1]['results']['zones']['limited_current_a']
    result={'status':'PASS-HOST','scope':'two independent local replay runs',
            'same_outcomes':True,'deterministic_fixtures_match':True,'m0':'M0_NOT_READY',
            'durations_s':[r['elapsed_s'] for r in runs],'outcomes':statuses[0]}
    (ROOT/'experiments/results/replay.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))


if __name__=='__main__':main()
