import os, time, sys
from os.path import isfile, join
import shutil

os.system('./build')

test_output = 'test_output'
tests = 'tests'
if len(sys.argv) == 2:
    tests = sys.argv[1]
try:
    shutil.rmtree(test_output)
except:
    pass
os.mkdir(test_output)
test_list = os.listdir(tests)
test_list.sort()
for f in test_list:
    abs_f = join(tests, f)
    if isfile(abs_f):
        if f[len(f) - len('.input'):] == '.input':
            fn = f[:len(f) - len('.input')]
            print(fn),
            os.system('python master.py < ' + abs_f + \
                      ' 2> ' + join(test_output, fn+'.err') + \
                      ' > ' + join(test_output, fn+'.output'))

            with open(join(test_output, fn+'.output')) as fi:
                out = fi.read()
            with open(join(tests, fn+'.output')) as fi:
                std = fi.read()
            if out.strip() == std.strip():
                print('correct')
            else:
                print('wrong')
            time.sleep(2)