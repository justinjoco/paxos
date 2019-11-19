package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sort"
	"time"
)

type Leader struct {
	pid            string
	replicas  	   []string
	acceptors 	   []string
	aliveAcceptors []string
	proposals      map[int]string
	ballotNum      int
	active         bool
	majorityNum    int
	slotNum		   int
}



func (self *Leader) Run(replicaLeaderChannel chan string) {
	self.active = false
	workerChannel := make(chan string)
	// spawn a scout
	go self.spawnScout(workerChannel)
	for {

		select {
		case replicaMsg := <-replicaLeaderChannel:
			messageSlice := strings.Split(replicaMsg, " ")
			self.slotNum, _ = strconv.Atoi(messageSlice[1])
			msgId := messageSlice[2]
			msg := messageSlice[3]
			proposal := msgId + " " + msg
			pprime := self.proposals[self.slotNum]
			if pprime == "" {
				self.proposals[self.slotNum] = proposal
				if self.active {
					go self.spawnCommander(workerChannel, self.slotNum, proposal)
				}
			}
		case workerMsg := <-workerChannel:
			messageSlice := strings.Split(workerMsg, ",")
			if messageSlice[0] == "adopted" {
				pvalues := messageSlice[2:]

				for _, pvalue := range pvalues {
					pvalSlice := strings.Split(pvalue, " ")
					found := false
				//	receivedBallot := pvalSlice[0]
					receivedSlot, _ := strconv.Atoi(pvalSlice[1])
					receivedProposal := pvalSlice[2] + " " + pvalSlice[3]
					for slot, _ := range self.proposals {
						if slot == receivedSlot {
							found = true
							break
						}
					}
					if !found{
						self.proposals[receivedSlot] = receivedProposal
					}

					for slot, proposal := range self.proposals {
						go self.spawnCommander(workerChannel, slot, proposal)
					}
				}

				self.active = true

			} else if messageSlice[0] == "preempted" {

				bprime, _ := strconv.Atoi(messageSlice[1]) //rprime in paper
				if bprime > self.ballotNum{
					self.active = false
					self.ballotNum = bprime + 1
					go self.spawnScout(workerChannel)
				}
				
			}
		}
	}
}



func (self *Leader) spawnScout(workerChannel chan string) {
	pvalues := make([]string, 0)
	alive := self.aliveAcceptors
	majorityNum := len(alive)/2 +1
	scoutAcceptorChannel := make(chan string)
	for _, processPort := range alive{
		go self.scoutTalkToAcceptor(processPort, scoutAcceptorChannel)

	}
	counter := 0
	for {
		select {
			case response := <- scoutAcceptorChannel:
				responseSlice := strings.Split(response, ",")
				keyWord := responseSlice[0]
				if keyWord == "p1b"{
		//			acceptorId := responseSlice[1]
					acceptorBallotInt, _ := strconv.Atoi(responseSlice[2])
					allAccepted := responseSlice[3:]
					if acceptorBallotInt == self.ballotNum {

						for _, receivedPval := range allAccepted {
							found := false
							for _, myPval := range pvalues {
								if myPval == receivedPval {
									found = true
									break
								}
							}
							if !found{
								pvalues = append(pvalues, receivedPval)
							}
						}

						counter += 1
						if counter >= majorityNum {

							
							pvalStr := ""
							for _, pval := range pvalues {
								pvalStr += "," + pval
							} 

							workerChannel <- "adopted," + strconv.Itoa(self.ballotNum) + pvalStr
							break
						}


					} else{
						workerChannel <- "preempted," + responseSlice[2]
						break
					}


				}
				
			default:
				continue
		}

	}

}

func (self *Leader) scoutTalkToAcceptor(processPort string, scoutAcceptorChannel chan string){
	acceptorConn, _ := net.Dial("tcp", "127.0.0.1:"+processPort)
	fmt.Fprintf(acceptorConn,"p1a," + self.pid + "," + strconv.Itoa(self.ballotNum))
	response, _ := bufio.NewReader(acceptorConn).ReadString('\n')
	scoutAcceptorChannel <- response
}

func (self *Leader) commTalkToAcceptor(processPort string, commAcceptorChannel chan string, slotNum int, proposal string){
	acceptorConn, _ := net.Dial("tcp", "127.0.0.1:"+processPort)
	fmt.Fprintf(acceptorConn,"p2a," + self.pid + "," + strconv.Itoa(self.ballotNum) + " " + strconv.Itoa(slotNum) + " " + proposal)
	response, _ := bufio.NewReader(acceptorConn).ReadString('\n')
	commAcceptorChannel <- response
}

func (self *Leader) spawnCommander(workerChannel chan string, slotNum int, proposal string) {
	alive := self.aliveAcceptors
	majorityNum := len(alive)/2 +1
	commAcceptorChannel := make(chan string)
	for _, processPort := range alive{
		go self.commTalkToAcceptor(processPort, commAcceptorChannel, slotNum, proposal)
	}
	counter := 0
	for {
		select {
			case response := <- commAcceptorChannel:
				responseSlice := strings.Split(response, ",")
				keyWord := responseSlice[0]
				if keyWord == "p2b"{
			//		acceptorId := responseSlice[1]
					acceptorBallotInt, _ := strconv.Atoi(responseSlice[2])
					if acceptorBallotInt == self.ballotNum {
						counter += 1
						if counter >= majorityNum {
							for _, replicaPort := range self.replicas {
								replicaConn, _ := net.Dial("tcp", "127.0.0.1:"+replicaPort)
								fmt.Fprintf(replicaConn,"decision," + strconv.Itoa(slotNum) + "," + proposal)
								bufio.NewReader(replicaConn).ReadString('\n')
							}
							break
						}
					} else{
						workerChannel <- "preempted," + responseSlice[2]
						break
					}


				}
				
			default:
				continue
		}

	}
	
}

func (self *Leader) Heartbeat() { //maintain alive list; calculate the majority

	for {

		tempAlive := make([]string, 0)

		for _, otherPort := range self.acceptors {

			if otherPort != self.pid {
				acceptorConn, err := net.Dial("tcp", "127.0.0.1:"+otherPort)
				if err != nil {
					continue
				}

				fmt.Fprintf(acceptorConn,"ping\n")
				response, _ := bufio.NewReader(acceptorConn).ReadString('\n')
				tempAlive = append(tempAlive, response)
			}

		}


		tempAlive = append(tempAlive, self.pid)
		sort.Strings(tempAlive)
		self.aliveAcceptors = tempAlive
		time.Sleep(1000 * time.Millisecond)
	}
}



