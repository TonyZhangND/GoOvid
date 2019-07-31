# GoOvid 

## What is this

GoOvid is the my own Go implementation of the 
[Ovid](https://www.usenix.org/system/files/conference/hotcloud16/hotcloud16_altinbuken.pdf), 
a system I worked on with Robbert van Renesse during my undergrad. Ovid is a framework 
for composing distributed services.

This project is a complete redesign of the original Ovid -- everything is architected from
the ground up. Ovid was implemented in C, using libuv as the messaging substrate. 
For the present work, I'm writing everything Go, to practice architecting my own system.

## Why this

1. Practice writing code and learn a new language 
   * The goal is design a system that is clean, extensible, well-documented, and leverages
    the features of the beautiful language that is Go. 
   * This is in fact the first time I am programing in Go
  
2. Framework for prototyping new protocols
   * I'm starting a PhD this year in distributed systems and Ovid is a super userful 
     framework for rapid prototyping of new protocols. Given any protocol, just program 
     the respective agents, and plug them into GoOvid with this simple, no-bake recipe. 

This directory has the following folders:

	chatroom:   A toy system used as a warm up exercise. It is a redesign
                of an undergrad project that I originally did in Python.
	commons:	Package containing common GoOvid definitions.
	configs:	Contains the GoOvid configuration files (.json), and Go package with the 
                utilities to parse those files.
	server:     Package containing the main server layer of GoOvid.

## TODO

1. A master script that parses that config file, and starts all servers
2. Implement a chat client that can send commands to a server
3. Doker-ize this baby