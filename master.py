"""
The master program for a GoOvid server.
Used to test the correctness of the server implementation 
"""

import os
import signal
import subprocess
import sys
import time
import platform
from socket import SOCK_STREAM, socket, AF_INET
from threading import Thread

address = 'localhost'
threads = {}  # ends up keeping track of who is alive
wait_ack = False


class ClientHandler(Thread):
    def __init__(self, index, address, port, process):
        Thread.__init__(self)
        self.index = index
        self.sock = socket(AF_INET, SOCK_STREAM)
        self.sock.connect((address, port))
        self.buffer = ""
        self.valid = True
        self.process = process

    def run(self):
        global threads, wait_ack
        while self.valid:
            if "\n" in self.buffer:
                (l, rest) = self.buffer.split("\n", 1)
                self.buffer = rest
                s = l.split()
                if s[0] == 'messages':
                    sys.stdout.write(l + '\n')
                    sys.stdout.flush()
                    wait_ack = False
                elif s[0] == 'alive':
                    sys.stdout.write(l + '\n')
                    sys.stdout.flush()
                    wait_ack = False
                else:
                    print("Invalid Response: " + l)
            else:
                try:
                    data = self.sock.recv(1024)
                    # sys.stderr.write(data)
                    self.buffer += data
                except:
                    # print(sys.exc_info())
                    self.valid = False
                    del threads[self.index]
                    self.sock.close()
                    break

    def kill(self):
        if self.valid:
            if platform.system() == 'Darwin':
                # MacOS
                self.send('crash\n')
            else:
                os.killpg(os.getpgid(self.process.pid), signal.SIGKILL)
            self.close()

    def send(self, s):
        if self.valid:
            self.sock.send(str(s) + '\n')

    def close(self):
        try:
            self.valid = False
            self.sock.close()
        except:
            pass


def kill(index):
    global wait_ack, threads
    wait = wait_ack
    while wait:
        time.sleep(0.01)
        wait = wait_ack
    pid = int(index)
    if pid >= 0:
        if pid not in threads:
            print('Master or testcase error!')
            return
        threads[pid].kill()


def send(index, data, set_wait_ack=False):
    global threads, wait_ack
    wait = wait_ack
    while wait:
        time.sleep(0.01)
        wait = wait_ack
    pid = int(index)
    if pid >= 0:
        if pid not in threads:
            print('Master or testcase error!')
            return
        if set_wait_ack:
            wait_ack = True
        threads[pid].send(data)
        return
    if set_wait_ack:
        wait_ack = True
    threads[pid].send(data)


def exit(force=False):
    global threads, wait_ack
    wait = wait_ack
    wait = wait and (not force)
    while wait:
        time.sleep(0.01)
        wait = wait_ack
    for k in threads:
        kill(k)
    subprocess.Popen(['./stopall'], stdout=open('/dev/null', 'w'), stderr=open('/dev/null', 'w'))
    time.sleep(0.1)
    if debug:
        print("Goodbye :)")
    sys.exit(0)


def timeout():
    time.sleep(120)
    print('Timeout!')
    exit(True)


def main(debug=False):
    global threads, wait_ack
    timeout_thread = Thread(target=timeout, args=())
    timeout_thread.setDaemon(True)
    timeout_thread.start()

    if debug:
        print("Master started")
    while True:
        line = ''
        try:
            line = sys.stdin.readline()
        except:  # keyboard exception, such as Ctrl+C/D
            exit(True)

        if line == '':  # end of a file
            exit()

        line = line.strip()  # remove trailing '\n'
        if line == '':  # prompt again if just whitespace
            continue

        if line == 'exit':  # exit when reading 'exit' command
            if debug:
                print("Received exit command. Terminating...")
            exit()

        sp1 = line.split(None, 1)
        sp2 = line.split()
        if len(sp1) != 2:  # validate input
            print("Invalid command: " + line)
            continue

        if sp1[0] == 'sleep':  # sleep command
            time.sleep(float(sp1[1]) / 1000)
            continue

        try:
            pid = int(sp2[0])  # first field is pid
        except ValueError:
            print("Invalid pid: " + sp2[0])
            exit(True)

        cmd = sp2[1]  # second field is command
        if cmd == 'start':
            try:
                port = int(sp2[3])
            except ValueError:
                print("Invalid port: " + sp2[3])
                exit(True)

            if debug:
                process = subprocess.Popen(['./process', str(pid), sp2[2], sp2[3]], preexec_fn=os.setsid)
            else:
                process = subprocess.Popen(['./process', str(pid), sp2[2], sp2[3]], stdout=open('/dev/null', 'w'),
                                           stderr=open('/dev/null', 'w'), preexec_fn=os.setsid)

            # sleep for a while to allow the process be ready
            time.sleep(3)

            # connect to the port of the pid
            handler = ClientHandler(pid, address, port, process)
            threads[pid] = handler
            handler.start()
        elif cmd == 'get' or cmd == 'alive':
            send(pid, sp1[1], set_wait_ack=True)
        elif cmd == 'broadcast':
            send(pid, sp1[1])
        elif cmd == 'crash':
            kill(pid)
            time.sleep(1)  # sleep for a bit so that crash is detected
        else:
            print("Invalid command: " + line)


if __name__ == '__main__':
    debug = False
    if len(sys.argv) > 1 and sys.argv[1] == 'debug':
        debug = True

    main(debug)