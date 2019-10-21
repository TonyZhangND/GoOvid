 #!/bin/bash
# This is a script that starts a paxos cluster

# Parse input
f=$1
nclients=$2
mode=$3
loss=$4

if [ $f -gt 49 ] 
then
    echo "Invalid args: Max f supported is 49. You used $f."
    exit 1
fi
if [ $nclients -gt 100 ] 
then
    echo "Invalid args: Max clients supported is 100. You used $nclients."
    exit 1
fi
if [ $mode != "script" ] && [ $mode != "manual" ] 
then
    echo "Invalid args: Mode is either 'script' or 'manual'."
    exit 1
fi
if [ $loss -gt 1 ] || [ $loss -lt 0 ] 
then
    echo "Invalid args: Loss is a percentage between 0-1."
    exit 1
fi


echo "Killing all running instances..."
./killall.sh  >/dev/null 2>&1

echo "Building GoOvid..."
./build

echo "Generating new configuration with"
echo "f=$f, nclients=$nclients, mode=$mode, networkloss=$loss"

python3 configs/paxos_generator.py $f $nclients $mode > configs/paxos.json

echo "Starting all boxes"
# TODO
# Generate list of box ids and start all of them





