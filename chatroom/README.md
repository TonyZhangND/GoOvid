# Chatroom

## Introduction 

To learn how to build distributed systems in Go, I am first implementing
this simple Chatroom project. I choose this project because this was the 
first lab for Cornell's *CS 5414: Principles of Distributed Systems*. 
This lab laid the groundwork for later projects, such as Three-Phase 
Commit and Paxos.

This Chatroom exercise will serve as the foundation for more complex
projects I envision, cumulating in a framework for composing distributed
services, i.e. Ovid.

## Specifications

I will follow the exact specifications in chatroom.pdf. The program should be 
able to run using master.py

Server id's always range from 0 - gridSize

### Port allocation

- Port numbers 1024 - 2999 are reserved for master - server
- Each server listens for connections on 3000 + physID
- Each server dials for connections on the range 3000 - 3000 + gridSize
  - Each server dials to servers whose physIDs is less than its own physID

### How connections are established

a connHandler registers its conn in the connTracker when it receives a ping
identifying the other party on the line

## Requirements

- Server can be compiled with the latest version of Go
- master.py should be run with Python 2. It is incompatible with Python 3