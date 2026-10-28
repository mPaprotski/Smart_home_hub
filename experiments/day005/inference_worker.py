"""Tiny local Python inference worker; synthetic logistic demo, NOT a Scene model."""
import json
import math
import sys


def infer(payload):
    if not isinstance(payload,dict) or set(payload)!={'brightness'}:raise ValueError('brightness required')
    x=payload['brightness']
    if isinstance(x,bool) or not isinstance(x,(int,float)) or not math.isfinite(x) or not 0<=x<=1:raise ValueError('invalid brightness')
    probability=1/(1+math.exp(-(8*x-4)))
    return {'demo_label':'BRIGHT_DEMO' if probability>=.5 else 'DARK_DEMO','probability':probability,'model':'synthetic-fixed-logistic-v1','is_project_scene_model':False}


def main():
    while True:
        raw=sys.stdin.buffer.readline(4097)
        if not raw:return
        if len(raw)>4096 or not raw.endswith(b'\n'):
            print(json.dumps({'error':'request_too_large'}),flush=True);return
        try:result=infer(json.loads(raw))
        except (ValueError,TypeError,json.JSONDecodeError):result={'error':'invalid_request'}
        print(json.dumps(result,allow_nan=False),flush=True)


if __name__=='__main__':main()
