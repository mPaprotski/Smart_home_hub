"""RAM-only synthetic edge analyzer, strict stream validation, optional live capture."""
import argparse
import json
import math
import os
import subprocess
import sys
import time
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import emit


def synthetic_frame(width=32,height=24):
    frame=[[(20,20,20) for _ in range(width)] for _ in range(height)]
    for y in range(height):
        for x in range(width):
            if y<2:frame[y][x]=(255,0,0)
            elif y>=height-2:frame[y][x]=(0,0,255)
            if x<2:frame[y][x]=(255,255,0)
            elif x>=width-2:frame[y][x]=(0,255,0)
    return frame


def edge_colors(frame,segments=2,depth=2):
    h=len(frame);w=len(frame[0]) if h else 0
    if isinstance(segments,bool) or not isinstance(segments,int) or segments<=0 or type(depth) is not int or depth<=0 or min(w,h)<=2*depth:
        raise ValueError('invalid geometry')
    if any(len(row)!=w for row in frame):raise ValueError('ragged frame')
    for row in frame:
        for p in row:
            if len(p)!=3 or any(isinstance(c,bool) or not isinstance(c,int) or c<0 or c>255 for c in p):raise ValueError('invalid RGB')
    result={}
    for side,length in [('top',w),('right',h),('bottom',w),('left',h)]:
        if segments>length-2*depth:raise ValueError('more segments than edge samples')
        colors=[]
        for segment in range(segments):
            start=depth+(length-2*depth)*segment//segments
            end=depth+(length-2*depth)*(segment+1)//segments
            sums=[0.,0.,0.];weights=0.
            for i in range(start,end):
                for d in range(depth):
                    weight=1/(d+1)
                    x,y=(i,d) if side=='top' else (w-1-d,i) if side=='right' else (i,h-1-d) if side=='bottom' else (d,i)
                    for c in range(3):sums[c]+=frame[y][x][c]*weight
                    weights+=weight
            colors.append([round(s/weights) for s in sums])
        result[side]=colors
    return result


def validate_packet(packet,*,epoch='synthetic-stream-a',zone='monitor.backlight',last_sequence=-1,count=8):
    if not isinstance(packet,dict):raise ValueError('object required')
    fields={'schema_version','stream_epoch','sequence','zone_id','mapping_version','colors'}
    if set(packet)!=fields:raise ValueError('unexpected/missing fields')
    if type(packet['schema_version']) is not int or packet['schema_version']!=1:raise ValueError('version')
    if packet['stream_epoch']!=epoch or packet['zone_id']!=zone:raise ValueError('binding/epoch')
    if type(packet['mapping_version']) is not int or packet['mapping_version']!=1:raise ValueError('mapping')
    if type(packet['sequence']) is not int or packet['sequence']<=last_sequence:raise ValueError('stale sequence')
    colors=packet['colors']
    if not isinstance(colors,list) or len(colors)!=count or len(colors)>1024:raise ValueError('length')
    for color in colors:
        if not isinstance(color,list) or len(color)!=3 or any(type(c) is not int or not 0<=c<=255 for c in color):raise ValueError('RGB')
    if len(json.dumps(packet,allow_nan=False).encode())>32768:raise ValueError('payload too large')
    return True


def capture_once():
    # Intentionally never serialize pixels or write screenshots. Only dimensions escape.
    try:
        from PIL import ImageGrab
        im=ImageGrab.grab()
        size=list(im.size);mode=im.mode
        im.close()
        print(json.dumps({'status':'PASS-HOST','dimensions':size,'mode':mode,'frame_persisted':False}))
    except Exception as exc:
        print(json.dumps({'status':'BLOCKED','reason':'display capture unavailable','error_type':type(exc).__name__}))


def probe(live=False):
    start=time.perf_counter();colors=edge_colors(synthetic_frame())
    expected={'top':[[255,0,0]]*2,'right':[[0,255,0]]*2,'bottom':[[0,0,255]]*2,'left':[[255,255,0]]*2}
    assert colors==expected
    packet={'schema_version':1,'stream_epoch':'synthetic-stream-a','sequence':1,'zone_id':'monitor.backlight','mapping_version':1,'colors':sum((colors[k] for k in ['top','right','bottom','left']),[])}
    assert validate_packet(packet)
    invalid=[]
    for key,value in [('schema_version',2),('schema_version',True),('stream_epoch','old'),('zone_id','desk.wall'),('sequence',-1),('sequence',1.5),('mapping_version',2),('colors',[[256,0,0]]*8)]:
        candidate=dict(packet);candidate[key]=value;invalid.append(candidate)
    invalid.append(dict(packet,screenshot='forbidden'))
    for candidate in invalid:
        try:validate_packet(candidate)
        except ValueError:continue
        raise AssertionError('invalid stream packet accepted')
    try:validate_packet(packet,last_sequence=1)
    except ValueError:pass
    else:raise AssertionError('duplicate sequence accepted')
    live_result={'status':'NOT_RUN','reason':'use --live-capture for a single RAM-only frame'}
    if live:
        try:
            r=subprocess.run([sys.executable,__file__,'--capture-child'],capture_output=True,text=True,timeout=4)
            live_result=json.loads(r.stdout) if r.returncode==0 else {'status':'BLOCKED','reason':'capture process error'}
        except (subprocess.TimeoutExpired,json.JSONDecodeError):live_result={'status':'BLOCKED','reason':'capture timeout/unavailable'}
    emit('D05-screen','PASS-SIM',['weighted four-edge golden fixture','ten invalid/replayed packet cases rejected'],
         synthetic_colors=colors,elapsed_ms=(time.perf_counter()-start)*1000,live_capture=live_result,
         other_targets={'Windows_x86_64':'BLOCKED_NO_TARGET','macOS_arm64':'BLOCKED_NO_TARGET','macOS_x86_64':'BLOCKED_NO_TARGET'},
         screenshots_saved=False)


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--live-capture',action='store_true');parser.add_argument('--capture-child',action='store_true')
    args=parser.parse_args()
    if args.capture_child:capture_once()
    else:probe(args.live_capture)
