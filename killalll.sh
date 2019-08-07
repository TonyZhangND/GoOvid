# This is a script that kills all chatroom processes

./stopall
pkill -9 -f 'ovid'
pkill -9 -f 'master.py'
pkill -9 -f 'grading.py'
