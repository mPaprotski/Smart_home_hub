"""Synthetic FFmpeg PCM decoding: no microphone, soundcard or audio files."""
import array
import math
import shutil
import subprocess
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from common import emit


def analyze(pcm, rate=48000, channels=2):
    if len(pcm) % (2 * channels):
        raise ValueError("partial PCM sample frame")
    data = array.array('h', pcm)
    if sys.byteorder != 'little': data.byteswap()
    left = [v / 32768 for v in data[::channels]]
    if not left: raise ValueError("empty PCM")
    rms = math.sqrt(sum(x*x for x in left) / len(left))
    peak = max(abs(x) for x in left)
    crossings = sum(a <= 0 < b for a, b in zip(left, left[1:]))
    active=[i for i,v in enumerate(left) if abs(v)>.001]
    active_frequency=None
    if len(active)>1:
        segment=left[max(0,active[0]-1):active[-1]+1]
        active_frequency=sum(a<=0<b for a,b in zip(segment,segment[1:]))*rate/len(segment)
    return {"sample_frames": len(left), "rms": rms, "peak": peak,
            "frequency_estimate_hz": crossings * rate / len(left),
            "active_frequency_estimate_hz": active_frequency}


def probe():
    if not shutil.which('ffmpeg'):
        return emit('D04-PCM', 'BLOCKED', [], reason='ffmpeg missing')
    def render(source):
        cmd=['ffmpeg','-nostdin','-v','error','-f','lavfi','-i',source,
             '-t','0.25','-ar','48000','-ac','2','-f','s16le','pipe:1']
        return subprocess.run(cmd, capture_output=True, timeout=10, check=True).stdout
    tone=analyze(render('sine=frequency=1000:sample_rate=48000'))
    silent=analyze(render('anullsrc=r=48000:cl=stereo'))
    assert tone['sample_frames']==12000
    assert 990<=tone['frequency_estimate_hz']<=1010
    assert 0.01<tone['rms']<0.2 and silent['rms']==0
    emit('D04-PCM','PASS-HOST',['FFmpeg synthetic source→PCM pipe',
         '48kHz stereo S16LE', 'tone sample count/frequency/RMS','silence'],
         tone=tone,silence=silent,physical_playback='BLOCKED_NO_AUDIO_DEVICE',
         persisted_pcm=False,scope='decoder/format only; not a production playback monitor')


if __name__=='__main__': probe()
