# GoOvid 

## What is this

GoOvid is the my own Go implementation of the 
[Ovid](https://www.usenix.org/system/files/conference/hotcloud16/hotcloud16_altinbuken.pdf), 
a system I worked on with Robbert van Renesse during my undergrad. Ovid is a framework 
for composing distributed services.

This project is a complete redesign of the original Ovid -- everything is architected from
the ground up. Ovid was implemented in C, using libuv as the messaging substrate. 
For the present work, I'm writing everything Go, to practice architecting my own system.

This directory has the following folders:

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

## Usage (In progress)

GoOvid/master.py is a tool that can be used to run GoOvid servers. 
It is a modification of GoOvid/chatroom/master.py, and one should refer 
to the documentation in that package, including the readme and chatroom.pdf, 
for more information. 

Currently, I am using this utility framework to test the correctness of the 
underlying server. It is subject to change as this project becomes more 
involved and a re-designed testing framework is needed.

To start the master, one runs 

```
python2 master.py <configFile> [-debug]
```

Below are the user commands for GoOvid/master.py, and their 
behavior.

|Input -> Master             |Master -> Server            |  Behavior                                    |
|----------------------	    |-------------------	         |-------------------------------------        |
| `<boxID> start`            | -                 	         | master starts the given box              	|
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



## Requirements

- Server can be compiled with the latest version of Go
- master.py and grading.py should be run with Python 2. It is explicitly incompatible with Python 3

## TODO

1. A master script that parses that config file, and starts all servers
2. Implement a chat client that can send commands to a server
3. Doker-ize this baby