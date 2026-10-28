"""Isolated PipeWire null sink: synthetic playback→monitor, never host microphone."""
import array
import math
import os
import select
import shutil
import subprocess
import sys
import tempfile
import threading
import time
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import ROOT, emit
from day004.pcm_probe import analyze


def probe():
    if not all(shutil.which(c) for c in ['pipewire','pw-cat','pw-link','wireplumber','dbus-daemon']):
        return emit('D04-PipeWire','BLOCKED',[],reason='PipeWire binaries missing')
    version=subprocess.run(['wireplumber','--version'],capture_output=True,text=True,timeout=5).stdout
    if '0.4.' not in version:
        return emit('D04-PipeWire','BLOCKED',[],reason='isolated recipe supports tested WirePlumber0.4 only; migrate config for0.5')
    processes=[];worker=None;stop=threading.Event()
    with tempfile.TemporaryDirectory(prefix='smarthome-pw-') as directory:
        env=dict(os.environ, XDG_RUNTIME_DIR=directory, PIPEWIRE_RUNTIME_DIR=directory,
                 PIPEWIRE_REMOTE='smarthome-spike')
        for key in ['PIPEWIRE_CONFIG_PREFIX','PIPEWIRE_CONFIG_DIR','PIPEWIRE_CONFIG_NAME']:
            env.pop(key,None)
        config=ROOT/'configs/pipewire-spike.conf'
        config_dir=Path(directory)/'wp-config'
        shutil.copytree('/usr/share/wireplumber',config_dir)
        (config_dir/'main.lua.d/60-spike-disable.lua').write_text('alsa_monitor.enabled = false\nv4l2_monitor.enabled = false\nlibcamera_monitor.enabled = false\n')
        (config_dir/'bluetooth.lua.d/60-spike-disable.lua').write_text('bluez_monitor.enabled = false\nbluez_midi_monitor.enabled = false\n')
        env['WIREPLUMBER_CONFIG_DIR']=str(config_dir)
        env['DBUS_SESSION_BUS_ADDRESS']='unix:path='+directory+'/session-bus'
        env['DBUS_SYSTEM_BUS_ADDRESS']='unix:path='+directory+'/no-system-bus'
        env['XDG_CONFIG_HOME']=directory+'/config'
        env['XDG_STATE_HOME']=directory+'/state'
        env['XDG_CACHE_HOME']=directory+'/cache'
        with tempfile.TemporaryFile() as log:
            server=subprocess.Popen(['pipewire','-c',str(config)],env=env,stdout=log,stderr=log)
            processes.append(server)
            def command(args,timeout=3):
                return subprocess.run(args,env=env,capture_output=True,text=True,timeout=timeout)
            def ports(flag):
                r=command(['pw-link',flag]);return r.stdout.splitlines() if r.returncode==0 else []
            def wait_ports(names,flag):
                until=time.monotonic()+3
                while time.monotonic()<until:
                    actual=ports(flag)
                    if all(n in actual for n in names):return
                    time.sleep(.05)
                raise RuntimeError('isolated graph ports did not become ready; actual=' + repr(actual))
            try:
                deadline=time.monotonic()+3
                while not (Path(directory)/'smarthome-spike').exists() and time.monotonic()<deadline:
                    if server.poll() is not None:break
                    time.sleep(.05)
                if server.poll() is not None or not (Path(directory)/'smarthome-spike').exists():
                    log.seek(0);message=log.read(4096).decode(errors='replace')
                    restricted='Operation not permitted' in message or 'Permission denied' in message
                    return emit('D04-PipeWire','BLOCKED' if restricted else 'FAIL',[],
                        reason='isolated daemon socket/permissions' if restricted else 'isolated daemon startup',
                        diagnostic_lines=[x[-180:] for x in message.splitlines()[-4:]],
                        host_audio_modified=False)
                bus=subprocess.Popen(['dbus-daemon','--session','--nofork','--address='+env['DBUS_SESSION_BUS_ADDRESS']],env=env,stdout=log,stderr=log)
                processes.append(bus)
                until=time.monotonic()+2
                while not (Path(directory)/'session-bus').exists() and time.monotonic()<until:time.sleep(.02)
                manager=subprocess.Popen(['wireplumber'],env=env,stdout=log,stderr=log)
                processes.append(manager)
                time.sleep(.3)
                play=subprocess.Popen(['pw-cat','--playback','--target','0','--rate','48000',
                     '--channels','2','--format','s16','--properties','{ node.name = hub_playback }','-'],
                     env=env,stdin=subprocess.PIPE,stdout=subprocess.DEVNULL,stderr=log)
                processes.append(play)
                def writer():
                    chunk=array.array('h',(int(6500*math.sin(2*math.pi*1000*i/48000)) for i in range(480) for _ in range(2)))
                    if sys.byteorder!='little':chunk.byteswap()
                    try:
                        while not stop.is_set(): play.stdin.write(chunk.tobytes());play.stdin.flush()
                    except (BrokenPipeError,OSError,ValueError):pass
                worker=threading.Thread(target=writer,daemon=True);worker.start()
                wait_ports(['hub_playback:output_FL','hub_playback:output_FR'],'-o')
                wait_ports(['hub_sink:playback_FL','hub_sink:playback_FR'],'-i')
                for channel in ['FL','FR']:
                    r=command(['pw-link',f'hub_playback:output_{channel}',f'hub_sink:playback_{channel}'])
                    if r.returncode:raise RuntimeError('failed to link playback to isolated sink')
                def capture(name):
                    rec=subprocess.Popen(['pw-cat','--record','--target','0','--rate','48000',
                        '--channels','2','--format','s16','--properties',f'{{ node.name = {name} }}','-'],
                        env=env,stdout=subprocess.PIPE,stderr=log)
                    processes.append(rec)
                    wait_ports([f'{name}:input_FL',f'{name}:input_FR'],'-i')
                    for channel in ['FL','FR']:
                        r=command(['pw-link',f'hub_sink:monitor_{channel}',f'{name}:input_{channel}'])
                        if r.returncode:raise RuntimeError('failed to link isolated sink monitor')
                    data=bytearray();until=time.monotonic()+1
                    while len(data)<48000*4//2 and time.monotonic()<until:
                        if select.select([rec.stdout],[],[],.1)[0]:
                            chunk=os.read(rec.stdout.fileno(),8192)
                            if not chunk:break
                            data.extend(chunk)
                    rec.terminate();rec.wait(timeout=3)
                    # --format s16 requests raw PCM. Reject unexpected container headers.
                    if bytes(data[:4]) in [b'.snd',b'RIFF']:raise RuntimeError('unexpected encoded PCM container')
                    result=analyze(bytes(data[:len(data)//4*4]))
                    if not .01<result['rms']<.3:raise RuntimeError('monitor missing synthetic tone')
                    if result['active_frequency_estimate_hz'] is None or not 980<result['active_frequency_estimate_hz']<1020:raise RuntimeError('monitor frequency mismatch')
                    return result
                first=capture('hub_capture_one')
                assert play.poll() is None and server.poll() is None
                second=capture('hub_capture_two')
                assert play.poll() is None
                emit('D04-PipeWire','PASS-HOST',['isolated null sink',
                     'synthetic playback→sink monitor→PCM', 'consumer stop/reconnect leaves playback alive'],
                     first_capture=first,second_capture=second,physical_audio='BLOCKED_NO_DEVICE',
                     bluetooth='BLOCKED_NO_ADAPTER',persisted_pcm=False,host_audio_modified=False)
            except (RuntimeError,subprocess.TimeoutExpired,AssertionError) as error:
                log.seek(0);tail=log.read(12000).decode(errors='replace').splitlines()[-6:]
                emit('D04-PipeWire','FAIL',[],reason=str(error),diagnostic_lines=tail,host_audio_modified=False)
            finally:
                stop.set()
                for process in reversed(processes):
                    if process.poll() is None:
                        process.terminate()
                        try:process.wait(timeout=3)
                        except subprocess.TimeoutExpired:process.kill();process.wait(timeout=3)
                    if process.stdin:
                        try:process.stdin.close()
                        except (BrokenPipeError,OSError):pass
                    if process.stdout:process.stdout.close()
                if worker:worker.join(timeout=1)


if __name__=='__main__':probe()
