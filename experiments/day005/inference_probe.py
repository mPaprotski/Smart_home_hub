"""Separate local worker/30 Hz synthetic loop with bounded queue and failure fallback."""
import json
import queue
import select
import subprocess
import sys
import threading
import time
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import emit


def probe():
    worker=subprocess.Popen([sys.executable,str(Path(__file__).with_name('inference_worker.py'))],stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL,text=True,bufsize=1)
    pending=queue.Queue(maxsize=1);results=[];stop=threading.Event()
    def inference():
        while not stop.is_set():
            try:job=pending.get(timeout=.1)
            except queue.Empty:continue
            try:
                worker.stdin.write(json.dumps({'brightness':job})+'\n');worker.stdin.flush()
                if not select.select([worker.stdout],[],[],.25)[0]:raise TimeoutError()
                result=json.loads(worker.stdout.readline())
                results.append(result)
            except (OSError,ValueError,TimeoutError):results.append({'error':'worker_unavailable'})
    thread=threading.Thread(target=inference,daemon=True);thread.start()
    start=time.monotonic();next_frame=start;last_infer=-10.;frames=0;requests=[];killed=False
    try:
        while time.monotonic()-start<2.2:
            now=time.monotonic()
            if now>=next_frame:
                frames+=1;next_frame+=1/30
            if now-start-last_infer>=1:
                try:pending.put_nowait(.8);requests.append(now-start);last_infer=now-start
                except queue.Full:pass
            if not killed and now-start>=1.25:
                worker.terminate();worker.wait(timeout=1);killed=True
            time.sleep(.001)
        assert frames>=55
        assert all(b-a>=.99 for a,b in zip(requests,requests[1:]))
        assert any('demo_label' in r for r in results)
        assert any(r.get('error')=='worker_unavailable' for r in results)
        emit('D05-inference','PASS-HOST',['separate stdlib Python worker',
             'bounded inference queue','at most one inference request per second',
             'worker failure does not stop synthetic colors loop'],
             duration_s=time.monotonic()-start,synthetic_frames=frames,
             request_times_s=requests,successful_inferences=sum('demo_label' in r for r in results),
             worker_failure_detected=True,is_project_scene_model=False,
             embedded_onnx_runtime='NOT_RUN_MISSING_DEPENDENCY')
    finally:
        stop.set()
        if worker.poll() is None:worker.terminate();worker.wait(timeout=2)
        thread.join(timeout=1)
        try:worker.stdin.close()
        except (BrokenPipeError,OSError):pass
        worker.stdout.close()


if __name__=='__main__':probe()
