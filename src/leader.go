package main

type Leader struct {
	pid       string
	replicas  []string
	acceptors []string
	ballotNum int
}

// functions:

func (self *Leader) Run(replicaLeaderChannel chan string) {

}

func (self *Leader) spawnScout() {

}

func (self *Leader) spawnCommander() {

}
