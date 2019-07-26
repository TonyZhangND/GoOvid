# Chatroom

## Introduction 

This project has two motivations. The first is to learn how to program in Go, and what 
better way to do it than to implement a distributed system! As a starter project, 
I am first implementing this simple Chatroom program. This was the first lab for 
Cornell's *CS 5414: Principles of Distributed Systems*. 
It laid the groundwork for later projects, such as Three-Phase Commit and Paxos.

Second, this Chatroom exercise will serve as the foundation for more complex
projects I envision, cumulating in a framework for composing distributed
services, something similar to [Ovid](https://www.usenix.org/system/files/conference/hotcloud16/hotcloud16_altinbuken.pdf), which was a system I worked on with 
Robbert van Renesse during my undergrad.

Notably, the original Ovid as implemented in C, using libuv as the messaging substrate. 
For the present work, I'm going to write my own layer in Go, mainly because libuv seems like
overkill, and I want practice working with my own sockets. 

The goal is design a system that is clean, extensible, well-documented, and leverages
the features of the beautiful language that is Go.

## Specifications

I will follow the exact specifications in chatroom.pdf. The program should be 
able to run using master.py

Each server has a physical id, and they always range from `0 - gridSize`.

Each server maintains a connection with the master. Links between servers
form a complete graph.

## Architecture

There are three logical components of a server
1. The **server** (server.go) houses the main logic of an Ovid server. In particular, it is the central location where actions are triggered by incoming messages.
2. The **linkManager** acts as a networking interface for an Ovid server. It manages all active links, and contains methods to query the state of the network.
3. The **link** represents a connection between two servers. It is responsible for maintaining and monitoring the health of the connection.

## Usage

Detailed usage and program behavior descriptions can be found in chatroom.pdf. 

In addition, the following command executes a test suite

```
python2 grading.py
```

that runs all the master scripts in the tests/ directory, comparing the actual output 
with the desired output for each test case.

Also, the following command

```
./killall.sh
```
kills all Ovid, master.py and grading.py processes.

### Port allocation

- Port numbers >= `10000` are reserved for master - server
- Ports `3000 to 3000 + gridSz` are used for inter-server links
- Each server listens for connections on `3000 + physID`
- Each server listens for connections on the range `3000 - 3000 + gridSz`
- Each server dials to servers whose physIDs is strictly less than its own physID

### Messaging format

Heartbeat pings are formatted as `ping [sender]`.

Other messages are formatted as `msg [sender] [contents]`.

## Requirements

- Server can be compiled with the latest version of Go
- master.py should be run with Python 2. It is explicitly incompatible with Python 3
