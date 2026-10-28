import copy
import json
import math
import subprocess
import sys
import unittest
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import ROOT
from day004.pcm_probe import analyze
from day005.screen_probe import edge_colors,synthetic_frame,validate_packet
from day005.inference_worker import infer
from day007.thermal_policy_probe import ThermalPolicy


class StreamTests(unittest.TestCase):
    def packet(self):
        return {'schema_version':1,'stream_epoch':'synthetic-stream-a','sequence':1,
                'zone_id':'monitor.backlight','mapping_version':1,'colors':[[1,2,3] for _ in range(8)]}
    def test_valid_packet(self):self.assertTrue(validate_packet(self.packet()))
    def test_bad_fields_and_bindings(self):
        for key,val in [('sequence',True),('sequence',0),('sequence',float('nan')),('stream_epoch','old'),('zone_id','other'),('mapping_version',2),('schema_version',True)]:
            with self.subTest(key=key,value=str(val)):
                p=self.packet();p[key]=val
                with self.assertRaises(ValueError):validate_packet(p,last_sequence=0)
    def test_duplicate_rejected(self):
        with self.assertRaises(ValueError):validate_packet(self.packet(),last_sequence=1)
    def test_extra_image_field_rejected(self):
        p=self.packet();p['screenshot']=[1,2]
        with self.assertRaises(ValueError):validate_packet(p)
    def test_rgb_ranges_types_and_lengths(self):
        for color in [[256,0,0],[-1,0,0],[True,0,0],[1.5,0,0],[0,0],[0,0,0,0]]:
            p=self.packet();p['colors'][0]=color
            with self.subTest(color=color),self.assertRaises(ValueError):validate_packet(p)
    def test_missing_pixel(self):
        p=self.packet();p['colors'].pop()
        with self.assertRaises(ValueError):validate_packet(p)


class AnalyzerTests(unittest.TestCase):
    def test_weighted_average_is_not_uniform_average(self):
        frame=synthetic_frame()
        for x in range(2,30):frame[0][x]=(240,0,0);frame[1][x]=(0,0,0)
        self.assertEqual(edge_colors(frame)['top'],[[160,0,0],[160,0,0]])
    def test_corner_exclusion(self):
        result=edge_colors(synthetic_frame())
        self.assertEqual(result['top'],[[255,0,0]]*2)
        self.assertEqual(result['left'],[[255,255,0]]*2)
    def test_small_ragged_and_invalid_geometry(self):
        for frame,args in [([],{}),([[(0,0,0)]],{}),(synthetic_frame(),{'segments':100}),(synthetic_frame(),{'depth':1.5})]:
            with self.subTest(args=args),self.assertRaises(ValueError):edge_colors(frame,**args)
    def test_pcm_rejects_partial_or_empty_frame(self):
        for data in [b'',b'123']:
            with self.assertRaises(ValueError):analyze(data)


class InferenceTests(unittest.TestCase):
    def test_monotonic_probabilities(self):
        probs=[infer({'brightness':x/10})['probability'] for x in range(11)]
        self.assertEqual(probs,sorted(probs))
        self.assertFalse(infer({'brightness':.2})['is_project_scene_model'])
    def test_bad_features(self):
        for x in [True,-1,2,None,float('nan'),float('inf')]:
            with self.subTest(x=str(x)),self.assertRaises(ValueError):infer({'brightness':x})
    def test_worker_bounds_and_invalid_json(self):
        script=ROOT/'experiments/day005/inference_worker.py'
        for content,error in [('garbage\n','invalid_request'),('a'*5000+'\n','request_too_large')]:
            r=subprocess.run([sys.executable,str(script)],input=content,text=True,capture_output=True,timeout=3,check=True)
            self.assertEqual(json.loads(r.stdout)['error'],error)
            self.assertNotIn(content.strip(),r.stdout)


class ThermalTests(unittest.TestCase):
    def config(self):return json.loads((ROOT/'configs/thermal-policy.example.json').read_text())
    def test_hysteresis_prevents_chatter(self):
        p=ThermalPolicy(self.config())
        self.assertEqual([p.update(v)['state'] for v in [50,49,48,47,46.9]],['WARM']*4+['NORMAL'])
    def test_invalid_input_keeps_cooling(self):
        p=ThermalPolicy(self.config())
        for t in [None,True,float('nan'),float('inf'),-100,1000,'65']:
            with self.subTest(t=str(t)):self.assertEqual(p.update(t)['fan'],'ON')
    def test_configuration_ordering(self):
        c=self.config();c['hot_c']=c['warm_c']
        with self.assertRaises(ValueError):ThermalPolicy(c)
    def test_extremes_skip_intermediate_states(self):
        p=ThermalPolicy(self.config())
        self.assertEqual(p.update(80)['state'],'CRITICAL')
        self.assertEqual(p.update(40)['state'],'NORMAL')


if __name__=='__main__':unittest.main()
