package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"os"
)

type Leader struct {
	pid       string
	replicas  []string
	acceptors []string
	proposals map[int]string
	ballotNum int
	active    bool
}

func (self *Leader) Run(replicaLeaderChannel chan string) {
	self.active = false
	workerChannel := make(chan string)
	// spawn a scout
	go self.spawnScout(workerChannel)
	for {
		// read go channel using select
		message, _ := reader.ReadString('\n')

		select {
		case replicaMsg := <-replicaLeaderChannel:
			messageSlice := strings.Split(replicaMsg, " ")
			slotNum := messageSlice[1]
			msgId := messageSlice[2]
			msg := messageSlice[3]
			proposal := msgId + " " + msg
			pprime := self.proposals[slotNum]
			if pprime == "" {
				self.proposals[slotNum] = proposal
				if self.active {
					go self.spawnCommander(workerChannel)
				}
			}
		case workerMsg := <-workerChannel:
			messageSlice := strings.Split(replicaMsg, " ")
			if messageSlice[0] == "adopted" {

			} else if messageSlice[0] == "preempted" {

			}
		}
	}
}

func (self *Leader) spawnScout(workerChannel chan string) {

}

func (self *Leader) spawnCommander(workerChannel chan string) {

	self.active = false
}
