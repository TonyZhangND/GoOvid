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

	agents:		Package containing the agents of GoOvid. 
	chatroom:	A toy system used as a warm up exercise. It is a redesign of an undergrad project that I 
			originally did in Python.
	commons:	Package containing common GoOvid definitions.
	configs:	Contains the GoOvid configuration files (.json), and Go package with the utilities to
			parse those files.
	server:		Package containing the main server layer of GoOvid.

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

Agents are the fundamental building blocks of GoOvid. Specifically, an **agent** is a 
self-contained state machine that transitions in response to messages it receives and 
may produce output messages for other agents. Services are implemented by an agent or a 
group of agents. For instance, a single KVS agent could implement a monolithic, non-fault-
tolerant key-value store; an Acceptor agent could implement a paxos acceptor; and together 
a group of Learner, Leader and Acceptor agents could implement a distributed, fault-tolerant
key-value store using the Paxos protocol.

The type of agents currently implemented in this repository includes:

* **Dummy** -- does nothing; used as a placeholder for server testing.
* **Chat** -- a tty service that broadcasts inputs from stdin, and prints incoming messages from other chat agents

To implement a new type of agent, one follows the below recipe:

1. Write a Go program containing a struct type definition of the new agent in the `agents` package, under GoOvid/agents/. This agent type must implement the `Agent` interface defined in GoOvid/agents/agentCommons.go.
2. Include the new agent type as an `AgentType` enum in GoOvid/agents/agentCommons.go.
3. Add a new switch case for the new agent type in the `NewAgent()` function in GoOvid/agents/agentCommons.go.
4. Add a new switch case for the new agent type in the `parseAgentObject()` function in GoOvid/configs/configParser.go.

### Boxes

A **box** is a container that is a single unit of failure in GoOvid. It is implemented user 
process running on the OS, and defined in the `server` package. Hence, a host machine can 
run multiple boxes, each of which can contain multiple agents. Boxes are completely 
transparent to the agents it contains -- an agent has no awareness of the box it is on, 
or of the other agents that are on the same box. Every box maintains a TCP connection 
with every other box to form a complete network graph.

### Configuration files

A **configuration** defines a system in GoOvid. It specifies the mapping of agents to 
boxes, the attributes of each agent, and the routing table of each agent.

Configuration files are written in the JSON language, which can then be read by GoOvid.
See GoOvid/configs/chat.json for an example. In the file, each agent is indexed by a unique
global identifier, of which the agent is not necessarily aware unless made explicit as an
atribute.

Each JSON agent object has the following characteristics:

* `type` -- A string describing the type of the agent, to be decoded by the parser
* `box` -- The box on which the agent resides. It is defined by it's external IP interface, i.e. an `"[IP]:[port]"` string, such as `127:0.0.1:10000` for an IPv4 address, and `[2601:646:2:df40:5924:f15a:a637:19ff]:5001` for IPv6. One need not worry about ambiguous representations of IP addresses. GoOvid will reduce the strings to their canonical address values for any comparison, such that the strings `127:0.0.1:10000` and `127:0.00.001:10000` refer to the same box, for instance. 
* `attrs` -- User-defined attributes for the particular agent. This can be an arbitrary JSON structure.
* `routes` -- The routing table of the agent. Each entry is defined by `<virtual dest> : { <physical dest> : <dest port> }`. 
  -  Since each agent is not necessarily aware of its physical ID or that of others, it sends messages to fixed virtual destinations. Each virtual destination points to the physical ID of the destination agent, and the port on which the server should deliver the message. 

### More on virtual and physical agent identifiers

A key design in Ovid is that there are two types of agent identifiers, virtual and physical (implementation wise, they are as of now the of same type `processID`). There are two arguments for this feature.

First, there is no reason why any agent show know, *a priori*, the identities of all other agents in the system -- this information is instead maintained at the level of the ovid framework. Usage wise, this allows for 'plug-and-play' agents. For instance, an agent could be programmed to send outgoing messages to virtual destination `2`. The physical agent that this virtual address `2` points to can be changed by just changing the routing table in the configuration, without any modifications to agent code.

Second, this allows for the system to *evolve dynamically* by only changing the GoOvid configuration, with the agent implemented as if the system is static. As an example, consider a client-server system, where there is one client agent of physical id `1` and one server agent of of physical id `2`. Suppose that the client uses the virtual id `200` for sending messages to a server, and the server uses `10` as its client receiving port. Thus the client agent routing table will contain the entry 

```
200: { 2, 10 }
```

This line means that when the client agent tries to send to virtual dest `200`, GoOvid delivers it to port `10` of agent `2`, which is the server agent. 

Now, we want to make the system fault tolerant by adding another server agent `3` that's a replica of `2`. Then the client agent needs to send every request to both servers. Instead of changing the client agent's code to do so, We can then use the routing table for *multiplexing*, by using two entries in the client's routing table

```
200: { 2, 10 }
200: { 3, 10 }
```

As a result, whenever the client agent tries to send to virtual dest `200`, GoOvid delivers it to both agents `2` and `3`. This feature allows for configurations to change dynamically in a running system.

### Starting a grid

A running system in GoOvid is called a **grid**. GoOvid parses a configuration file and automatically contructs a grid according to that configuration, agents, boxes and all.

To build GoOvid, run 

```
./build
```

To start a grid, run the command

```
./ovid [-debug] [path/to/configfile] [box]
```

and ovid will start all agents residing in the box. Note that all command line flags must 
be placed before positional arguments. 

To quickly kill all Ovid processes, run the command

```
./killall.sh
```

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