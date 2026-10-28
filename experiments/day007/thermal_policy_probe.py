"""Pure state-machine spike; never writes sysfs or drives a physical fan."""
import json
import math
import sys
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import ROOT, emit


class ThermalPolicy:
    labels=['NORMAL','WARM','HOT','CRITICAL']
    def __init__(self,config):
        self.thresholds=[config[k] for k in ['warm_c','hot_c','critical_c']]
        h=config['hysteresis_c']
        if any(isinstance(x,bool) or not isinstance(x,(int,float)) or not math.isfinite(x) for x in self.thresholds+[h]):raise ValueError('finite thresholds required')
        if not 0<h or not self.thresholds[0]<self.thresholds[1]<self.thresholds[2]:raise ValueError('threshold ordering/hysteresis')
        if h>=min(b-a for a,b in zip(self.thresholds,self.thresholds[1:])):raise ValueError('hysteresis crosses state boundaries')
        self.hysteresis=h;self.index=0
    def update(self,temp):
        if isinstance(temp,bool) or not isinstance(temp,(int,float)) or not math.isfinite(temp) or not -40<=temp<=150:
            return {'state':self.labels[self.index],'quality':'ERROR','fan':'ON'}
        while self.index<3 and temp>=self.thresholds[self.index]:self.index+=1
        while self.index>0 and temp<self.thresholds[self.index-1]-self.hysteresis:self.index-=1
        return {'state':self.labels[self.index],'quality':'VALID','fan':'OFF' if self.index==0 else 'ON'}


def probe():
    conf=json.loads((ROOT/'configs/thermal-policy.example.json').read_text());p=ThermalPolicy(conf)
    seq=[45,50,49.9,48,47,46.9,65,64.9,62,61.9,75,74,72,71.9]
    expected=['NORMAL','WARM','WARM','WARM','WARM','NORMAL','HOT','HOT','HOT','WARM','CRITICAL','CRITICAL','CRITICAL','HOT']
    actual=[p.update(t)['state'] for t in seq];assert actual==expected,(actual,expected)
    for bad in [None,float('nan'),float('inf'),True,'50',200]:assert p.update(bad)['fan']=='ON'
    emit('D07-thermal','PASS-SIM',['four states','rising/falling hysteresis boundaries',
         'six invalid sensor cases fail to fan ON'],synthetic_temperatures_c=seq,
         states=actual,physical_fan='BLOCKED_NO_PI',hardware_sysfs_writes=False,
         defaults='proposed only; calibration pending')


if __name__=='__main__':probe()
