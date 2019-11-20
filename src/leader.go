package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Leader struct {
	pid       string
	replicas  []string
	acceptors []string
	proposals   map[int]string
	ballotNum   int
	active      bool
	majorityNum int
	slotNum     int
}

func (self *Leader) Run(replicaLeaderChannel chan string) {
	self.active = false
	workerChannel := make(chan string)

	go self.spawnScout(workerChannel)
	for {

		select {
		case replicaMsg := <-replicaLeaderChannel:
			fmt.Println("RECEIVED FROM REPLICA:" + replicaMsg)
	
			messageSlice := strings.Split(replicaMsg, " ")
			if messageSlice[0] == "propose" { // If message from the replica is propose
				self.slotNum, _ = strconv.Atoi(messageSlice[1])
				msgId := messageSlice[2]
				msg := messageSlice[3]
				proposal := msgId + " " + msg
			
				fmt.Println(self.proposals[self.slotNum])
				
				self.proposals[self.slotNum] = proposal
				fmt.Println(self.proposals)
			
				go self.spawnCommander(workerChannel, self.slotNum, proposal)

			} else if messageSlice[0] == "catchup" {
				fmt.Println("SEND CATCHUP")
				for _, replicaPort := range self.replicas {
					replicaConn, err := net.Dial("tcp", "127.0.0.1:"+replicaPort)
					if err != nil{
						fmt.Println(err)
						replicaLeaderChannel <- ""
						continue
					}
					fmt.Fprintf(replicaConn, "catchup\n")
					response, _ := bufio.NewReader(replicaConn).ReadString('\n')
					response = strings.TrimSuffix(response, "\n")
					replicaLeaderChannel <- response
					replicaConn.Close()
				}

			}

		case workerMsg := <-workerChannel:
			fmt.Println("RECEIVED FROM WORKER:" + workerMsg)
			messageSlice := strings.Split(workerMsg, ",")

			if messageSlice[0] == "adopted" {
			
				pvalues := messageSlice[2:]
				fmt.Println(pvalues)
				self.active = true
				
				

			} else if messageSlice[0] == "preempted" {
				fmt.Println("PREEMPTED")
				bprime, _ := strconv.Atoi(messageSlice[1]) //rprime in paper
				if bprime > self.ballotNum {
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
	
	fmt.Println("PID:" + self.pid + " SCOUT SPAWNED")
	majorityNum := len(self.acceptors)/2 + 1
	scoutAcceptorChannel := make(chan string)
	if crashStage == "p1a" {
		if len(crashAfterSentTo) == 0 {
			os.Exit(1)
		}
		for _, acceptorId := range crashAfterSentTo {
			acceptorIdInt, _ := strconv.Atoi(acceptorId)
			acceptorPort := strconv.Itoa(acceptorIdInt + 20000)
			acceptorConn, _ := net.Dial("tcp", "127.0.0.1:"+acceptorPort)
			fmt.Fprintf(acceptorConn, "p1a,"+self.pid+","+strconv.Itoa(self.ballotNum)+"\n")
			acceptorConn.Close()
		}
		os.Exit(1)
	} else {
	
		for _, acceptorPort := range self.acceptors {
			go self.scoutTalkToAcceptor(acceptorPort, scoutAcceptorChannel)

		}
	}
	counter := 0
	for {
		select {
		case response := <-scoutAcceptorChannel:
			fmt.Println(response)
			responseSlice := strings.Split(response, ",")
			keyWord := responseSlice[0]
			if keyWord == "p1b" {
			
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
						if !found {
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
						counter = 0
						break
					}

				} else {
					workerChannel <- "preempted," + responseSlice[2]
					break
				}

			}
		}

	}

}

func (self *Leader) spawnCommander(workerChannel chan string, slotNum int, proposal string) {
	// alive := self.aliveAcceptors
	fmt.Println("PID:" + self.pid + " COMMANDER SPAWNED")
	majorityNum := len(self.acceptors)/2 + 1
	commAcceptorChannel := make(chan string)
	if crashStage == "p2a" {
		if len(crashAfterSentTo) == 0 {
			os.Exit(1)
		} else {
			for _, acceptorId := range crashAfterSentTo {
				acceptorIdInt, _ := strconv.Atoi(acceptorId)
				acceptorPort := strconv.Itoa(acceptorIdInt + 20000)
				acceptorConn, _ := net.Dial("tcp", "127.0.0.1:"+acceptorPort)
				fmt.Fprintf(acceptorConn, "p2a,"+self.pid+","+strconv.Itoa(self.ballotNum)+" "+strconv.Itoa(slotNum)+" "+proposal+"\n")
				acceptorConn.Close()
			}
			os.Exit(1)
		}
	} else {
		for _, acceptorPort := range self.acceptors {
			go self.commTalkToAcceptor(acceptorPort, commAcceptorChannel, slotNum, proposal)
		}
	}

	counter := 0
	for {
		select {
		case response := <-commAcceptorChannel:
			fmt.Println(response)
			responseSlice := strings.Split(response, ",")
			keyWord := responseSlice[0]
			if keyWord == "p2b" {
				//		acceptorId := responseSlice[1]
				acceptorBallotInt, _ := strconv.Atoi(responseSlice[2])
				fmt.Println("ACCEPTOR BALLOT INT:"+ string(responseSlice[2]))
				if acceptorBallotInt == self.ballotNum {
					counter += 1
					fmt.Println(counter)
					if counter >= majorityNum {
						counter = 0
						if crashStage == "decision" {
							if len(crashAfterSentTo) == 0 {
								os.Exit(1)
							} else {
								for _, replicaId := range crashAfterSentTo {
									replicaIdInt, _ := strconv.Atoi(replicaId)
									replicaPort := strconv.Itoa(replicaIdInt + 20100)
									replicaConn, _ := net.Dial("tcp", "127.0.0.1:"+replicaPort)
									fmt.Fprintf(replicaConn, "decision,"+strconv.Itoa(slotNum)+","+proposal+"\n")
									replicaConn.Close()

								}
								os.Exit(1)
							}
						}
						fmt.Println("DECISION SENDING")
						fmt.Println(self.proposals)
						
						for _, replicaPort := range self.replicas {
							replicaConn, err := net.Dial("tcp", "127.0.0.1:"+replicaPort)
							if err != nil{
								fmt.Println(err)
								continue
							}
							fmt.Fprintf(replicaConn, "decision "+strconv.Itoa(slotNum)+" "+proposal + "\n")
						

							replicaConn.Close()
						}
						workerChannel <- ""
						break
					}
				} else {

					workerChannel <- "preempted," + responseSlice[2]
					break
				}

			}

		}

	}

}


func (self *Leader) scoutTalkToAcceptor(processPort string, scoutAcceptorChannel chan string) {
	fmt.Println("Scout is talking to this acceptor: " + processPort)
	acceptorConn, err := net.Dial("tcp", "127.0.0.1:"+processPort)
	if err != nil {
		fmt.Println(err)
		scoutAcceptorChannel <- ""
		return
	}
	fmt.Println("Sending p1a")
	fmt.Fprintf(acceptorConn, "p1a,"+self.pid+","+strconv.Itoa(self.ballotNum)+"\n")
	response, _ := bufio.NewReader(acceptorConn).ReadString('\n')
	acceptorConn.Close()
	scoutAcceptorChannel <- response
}

func (self *Leader) commTalkToAcceptor(processPort string, commAcceptorChannel chan string, slotNum int, proposal string) {
	acceptorConn, err := net.Dial("tcp", "127.0.0.1:"+processPort)
	if err != nil {
		fmt.Println(err)
		commAcceptorChannel <- ""
		return
	}
	fmt.Println("Sending p2a")
	fmt.Fprintf(acceptorConn, "p2a,"+self.pid+","+strconv.Itoa(self.ballotNum)+" "+strconv.Itoa(slotNum)+" "+proposal+"\n")
	response, _ := bufio.NewReader(acceptorConn).ReadString('\n')
	acceptorConn.Close()
	commAcceptorChannel <- response
}



