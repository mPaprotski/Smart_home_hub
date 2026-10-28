"""Hardware evidence gate only. Never pretends synthetic outputs are Pi tests."""
import importlib.util
import sys
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import ROOT,emit


def probe():
    spec=importlib.util.spec_from_file_location('pi_preflight',ROOT/'deploy/bootstrap/pi_preflight.py')
    mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
    target=mod.inspect()
    cases={f'D08-{i:02d}':{'status':'BLOCKED','reason':'target Pi/LED/audio fixture not available for this session'} for i in range(1,11)}
    emit('D08-HIL-gate','BLOCKED',[],target=target,hardware_cases=cases,
         gpio_writes=False,reason='user confirmed Pi absent or not configured; no SSH target')


if __name__=='__main__':probe()
