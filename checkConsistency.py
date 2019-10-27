# Checks dumps for consistency issues

import glob


outputFiles = glob.glob('tmp/replica*.output')
outputFiles.sort()

# number of lines in the longest file
max_lines =  max([sum(1 for line in open(f)) for f in outputFiles])

file_lines = []
for name in outputFiles:
    f = open(name)
    file_lines.append(f.readlines())

for i in range(max_lines):
    s = set()
    for f in file_lines:
        if i < len(f):
            s.add(f[i])
    if len(s) > 1:
        print(f"Inconsistency detected in line {i+1}")
        lines = [fl[i].strip() for fl in file_lines]
        print(",\n".join(lines))
        exit(1)
print("All good :)")