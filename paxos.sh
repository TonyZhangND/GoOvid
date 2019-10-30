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

if (( $(echo "$loss > 1" |bc -l) )) || (( $(echo "$loss < 0" |bc -l) ))
then
    echo "Invalid args: Loss is a percentage between 0-1."
    exit 1
fi


echo "Killing all running instances and cleaning directory..."
./killall.sh  >/dev/null 2>&1
rm -f nohup.out
rm tmp/box*.log
rm tmp/replica*.output

echo "Generating new configuration with"
echo "f=$f, nclients=$nclients, mode=$mode, networkloss=$loss"

python3 configs/paxos_generator.py $f $nclients $mode > configs/paxos.json

echo "Starting all boxes"

# Start replica boxes in background
replicaID=1
if [ $mode != "manual" ]
then
    while [ $replicaID -lt $(( $f*2 + 2 )) ] 
    do
        let port=5000+$replicaID
        box="127.0.0.1:$port"
        let replicaID++
        nohup ./ovid -loss=$4 configs/paxos.json $box &
        # nohup ./ovid -loss=$4 configs/paxos.json $box &
    done
fi

if [ $mode != "manual" ]
then
    # Start client boxes in background
    clientID=100
    while [ $clientID -lt $(( 100 + $nclients )) ] 
    do
        let port=8000+$clientID
        box="127.0.0.1:$port"
        let clientID++
        nohup ./ovid -log configs/paxos.json $box &
        # nohup ./ovid configs/paxos.json $box &
    done
fi
disown

# Start special controller box in foreground
./ovid -debug configs/paxos.json 127.0.0.1:9999
./killall.sh






