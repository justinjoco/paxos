# Paxos in Go
Authors: Dilip Reddy (dr589), Justin Joco (jaj263), Zhilong Li (zl242)

Slip days used(this project): 0  

Slip days used(total):

* Dilip (dr589): 4

* Justin (jaj263): 2

* Zhilong Li (zl252): 2

## General Overview
We implemented Multi-Paxos, a distributed consensus algorithm, by following and maintaining the specs and invariants described in "Paxos Made Moderately Complex" by Robbert Van Reneese and Deniz Altinbuken, though our code diverges from the paper's pseudocode. Each process performs all of the following 5 roles: Replica, Leader, Acceptor, Commander, and Scout.

On a high level, what each role does is the following:

- Replica: Client-facing role that prompts the leader to propose a command to the other servers at a specific slot number, performs the decisions agreed by the majority, and responds accordingly to client.
- Leader: Role that spawns scouts and commanders to persuade the acceptors to adopt a given ballot number at a given slot number, then accept a given proposal
- Acceptor: Passive role that responds to a leader's scouts and commanders with its current ballot number and list of accepted proposals; adopts a ballot number if it's greater than its current ballot number
- Scout: Spawned by leader to persuade the acceptors of other servers to adopt a given ballot number
- Commander: Spawned by leader to tell the acceptors of other servers to accept a proposal with a ballot number at a specific slot number


### Automated testing 
Within the root directory of this project directory, run your test script with a master process here.
For more inpromptu script testing, run `python master.py < <testFile>.input`, which will return the test output to stdout.

### Non-deterministic/impromptu testing
Assume that you are in the root directory of this project using a terminal:
* To build, run `./build` in order to create a binary Go file, which will be created in a /bin/ folder.

For each worker process you want to run, open a new window and go to this project directory and do the following:
* Run the Go program by running `./process <id> <n> <portNum>`, which runs the executable file in the /bin/ folder from earlier.
* Use 'CTRL C' to end this specific process.

To stop all Go processes from running and destroy the DT logs, run `./stopall`.

Use master.py to send your commands directly to the distributed processes.

### OS Testing Environment
Our group used OSX Mojave 10.14.x and Ubuntu Linux to compile, run, and test this project.






