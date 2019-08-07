# GoOvid 

## What is this

GoOvid is the my own Go implementation of the 
[Ovid](https://www.usenix.org/system/files/conference/hotcloud16/hotcloud16_altinbuken.pdf), 
a system I worked on with Robbert van Renesse during my undergrad. Ovid is a 
contained-based framework for composing distributed services.

This project is a complete redesign of the original Ovid -- everything is architected from
the ground up. Ovid was implemented in C, using libuv as the messaging substrate. 
For the present work, I'm writing everything Go, to practice architecting my own system.

This directory has the following folders:
     agents:   Package containing the agents of GoOvid. 
	chatroom: A toy system used as a warm up exercise. It is a redesign of an undergrad project that I
               originally did in Python.
	commons:  Package containing common GoOvid definitions.
	configs:  Contains the GoOvid configuration files (.json), and Go package with the utilities to
               parse those files.
	server:   Package containing the main server layer of GoOvid.

## Why this

1. Practice writing code and learn a new language 
   * The goal is design a system that is clean, extensible, well-documented, and leverages
    the features of the beautiful language that is Go. 
   * This is in fact the first time I am programing in Go
  
2. Framework for prototyping new protocols
   * I'm starting a PhD this year in distributed systems and Ovid is a super userful 
     framework for rapid prototyping of new protocols. Given any protocol, just program 
     the respective agents, and plug them into GoOvid with this simple, no-bake recipe. 

## Usage 

### Agents

Agents are the fundamental building blocks of GoOvid. Specifically, an agent is a 
self-contained state machine that transitions in response to messages it receives and 
may produce output messages for other agents. Services are implemented by an agent or a 
group of agents. For instance, a KVS agent could implement a monolithic key-value store;
an Acceptor agent could implement a paxos acceptor; and together a group of Learner, Leader
and Acceptor agents could implement Paxos.

The type of agents currently implemented in this repository includes:

* **Dummy** -- does nothing; used as a placeholder for server testing.
* **Chat** -- a tty service that broadcasts inputs from stdin, and prints incoming messages from other chat agents

To implement a new type of agent, one follows the below recipe:

1. Write a Go program containing a struct type definition of the new agent in the `agents` package, under GoOvid/agents/. This agent type must implement the `Agent` interface defined in GoOvid/agents/agentCommons.go.
2. Include the new agent type as an `AgentType` enum in GoOvid/agents/agentCommons.go.
3. Add a new switch case for the new agent type in the `NewAgent()` function in GoOvid/agents/agentCommons.go.
4. Add a new switch case for the new agent type in the `parseAgentObject()` function in GoOvid/configs/configParser.go.

### Boxes

### Configuration files

### Starting a grid


## Testing the servers

GoOvid/master.py is a tool that can be used to test the correctness of GoOvid's server
layer, that is, the boxes that agents reside it. 
It is a modification of GoOvid/chatroom/master.py, and one should refer 
to the documentation in that package, including the readme and chatroom.pdf, 
for more information. 

To start the master, one runs 

```
python2 master.py <configFile> [debug]
```

One can then enter commands into the master.py program to direct the server to perform
certain primitives. Below are the user commands for GoOvid/master.py, and their behavior.

|Input -> Master             |Master -> Server            |  Behavior                                    |
|----------------------	    |-------------------	         |-------------------------------------        |
| `<boxID> start <port>`     | -                 	         | master starts the given box with `./process <configFile> <boxID> <port>`|
| `exit`                     | -                 	         | master calls `./stopall` then exits       	|
| `sleep <n>`                | -                 	         | master sleeps for `n` milliseconds          |
| `<boxID> crash`            | -                  	    | master crashes the given box              	|
| `<boxID> get`              | `get`                       | the receiver responds to the master with its message log |
| `<boxID> alive`           |  `alive`                     | the receiver responds to the master with the id of all boxes it thinks are alive, including itself |
| `<boxID> broadcast <msg>`  |  `broadcast <msg>`          | the receiver broadcasts the given message to all boxes alive, including itself |

Below are the responses that servers should return to the master for the 
respective commands.

|Server ->  Master           | Description |
|----------------------	    |-------------------	         |
| `alive <id1>,<id2>,...`    | a box asked to return all alive boxes responds by giving a list of the box ids in ascending order  | 
| `messages <m1>,<m2>,...`   | a box asked to return its messages responds by giving a list of all messages it has received in FIFO order |

GoOvid/grading.py is a program built on top of master.py that runs a battery of tests 
against the GoOvid server layer, and verifies the result. To run it, one does

```
python2 grading.py
```

Note that grading.py uses the configuration tests/test.json.

## Requirements

- Server can be compiled with the latest version of Go
- master.py and grading.py should be run with Python 2. It is explicitly incompatible with Python 3

## TODO

1. Doker-ize this baby
2. Framework for testing agents