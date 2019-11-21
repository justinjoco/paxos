package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Leader struct {
	pid         string
	replicas    map[int]string
	acceptors   map[int]string
	proposals   map[int]string
	ballotNum   int
	active      bool
	majorityNum int
	slotNum     int
}

//Spawns scout on proposal, spawns commander on adopted msg from scout
func (self *Leader) Run(replicaLeaderChannel chan string) {
	self.active = false
	workerChannel := make(chan string)

	for {

		select {

			case replicaMsg := <-replicaLeaderChannel:

				messageSlice := strings.Split(replicaMsg, " ")
				if messageSlice[0] == "propose" { // If message from the replica is propose
					self.slotNum, _ = strconv.Atoi(messageSlice[1])
					msgId := messageSlice[2]
					msg := messageSlice[3]
					proposal := msgId + " " + msg		
					self.proposals[self.slotNum] = proposal

					go self.SpawnScout(workerChannel)
				

				} else if messageSlice[0] == "catchup" {
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
				
					}

				}

			case workerMsg := <-workerChannel:
				messageSlice := strings.Split(workerMsg, ",")

				if messageSlice[0] == "adopted" {

					go self.SpawnCommander(workerChannel, self.slotNum, self.proposals[self.slotNum])
					

				} else if messageSlice[0] == "preempted" {
					bprime, _ := strconv.Atoi(messageSlice[1]) //rprime in paper
					if bprime > self.ballotNum {
						self.ballotNum = bprime + 1
						go self.SpawnScout(workerChannel)
					}

				}
		}
	}
}



//Spawn a scout to get other acceptors to accept ballot number, tells leader if ballot is preempted
func (self *Leader) SpawnScout(workerChannel chan string) {
	pvalues := make([]string, 0)
	majorityNum := len(self.acceptors)/2 + 1
	scoutAcceptorChannel := make(chan string)
	if crashStage == "p1a" {
		if len(crashAfterSentTo) == 0 {
			os.Exit(1)
		}
		for _, acceptorId := range crashAfterSentTo {
			acceptorIdInt, _ := strconv.Atoi(acceptorId)
			acceptorPort := self.acceptors[acceptorIdInt]
			acceptorConn, _ := net.Dial("tcp", "127.0.0.1:"+acceptorPort)
			fmt.Fprintf(acceptorConn, "p1a,"+self.pid+","+strconv.Itoa(self.ballotNum)+"\n")
			
		}
		os.Exit(1)
	} 
	
	retry := true


	//Retry if we don't get a majority of acceptor responses and don't receive preempted
	for {
		for _, acceptorPort := range self.acceptors {
			go self.ScoutTalkToAcceptor(acceptorPort, scoutAcceptorChannel)

		}	
		counter := 0
		numResponses := 0
		for numResponses < len(self.acceptors){
			select {
			case response := <-scoutAcceptorChannel:
				numResponses +=1
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
							retry = true
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
						retry = false
						break
					}

				}
			}

		}
		if !retry {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}

}
//Spawn a commander to get acceptor to accept proposal
func (self *Leader) SpawnCommander(workerChannel chan string, slotNum int, proposal string) {
	majorityNum := len(self.acceptors)/2 + 1
	commAcceptorChannel := make(chan string)
	if crashStage == "p2a" {

		//crashing after sending p2a to no one
		if len(crashAfterSentTo) == 0 {
			os.Exit(1)
		} else {
			for _, acceptorId := range crashAfterSentTo {
				acceptorIdInt, _ := strconv.Atoi(acceptorId)
				acceptorPort := self.acceptors[acceptorIdInt]
				acceptorConn, _ := net.Dial("tcp", "127.0.0.1:"+acceptorPort)
				fmt.Fprintf(acceptorConn, "p2a,"+self.pid+","+strconv.Itoa(self.ballotNum)+" "+strconv.Itoa(slotNum)+" "+proposal+"\n")
			}
			//crashing after sending p2a 
			os.Exit(1)
		}
	} 

	//Retry if we don't get a majority of acceptor responses and don't receive preempted
	retry := true
	for {
		// sending p2A to acceptors
		for _, acceptorPort := range self.acceptors {
			go self.CommTalkToAcceptor(acceptorPort, commAcceptorChannel, slotNum, proposal)
		}

		counter := 0
		numResponses := 0
		for numResponses < len(self.acceptors){
			select {
			//Get response from acceptor
			case response := <-commAcceptorChannel:
				numResponses += 1
				responseSlice := strings.Split(response, ",")
				keyWord := responseSlice[0]
				if keyWord == "p2b" {
					acceptorBallotInt, _ := strconv.Atoi(responseSlice[2])
					if acceptorBallotInt == self.ballotNum {
						counter += 1
						if counter >= majorityNum {
							counter = 0
							//Crashing after sending decision to no one
							if crashStage == "decision" {
								if len(crashAfterSentTo) == 0 {
									os.Exit(1)
								} else {
									for _, replicaId := range crashAfterSentTo {
										replicaIdInt, _ := strconv.Atoi(replicaId)
										replicaPort := self.replicas[replicaIdInt]
										replicaConn, _ := net.Dial("tcp", "127.0.0.1:"+replicaPort)
										fmt.Fprintf(replicaConn, "decision,"+strconv.Itoa(slotNum)+","+proposal+"\n")
									}
									os.Exit(1)
								}
							}
							retry = false
							//Send decision to all
							for _, replicaPort := range self.replicas {
								replicaConn, err := net.Dial("tcp", "127.0.0.1:"+replicaPort)
								if err != nil{
									fmt.Println(err)
									continue
								}
								fmt.Fprintf(replicaConn, "decision "+strconv.Itoa(slotNum)+" "+proposal + "\n")
							}
							workerChannel <- ""
							break
						}
					} else {
						workerChannel <- "preempted," + responseSlice[2]
						retry = false
						break
					}
				}

			}

		}
		
		if !retry {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}

}

//Scout talks to one acceptor; tells leader the acceptor's response
func (self *Leader) ScoutTalkToAcceptor(processPort string, scoutAcceptorChannel chan string) {

	acceptorConn, err := net.Dial("tcp", "127.0.0.1:"+processPort)
	if err != nil {
		fmt.Println(err)
		scoutAcceptorChannel <- ""
		return
	}
	fmt.Fprintf(acceptorConn, "p1a,"+self.pid+","+strconv.Itoa(self.ballotNum)+"\n")
	response, _ := bufio.NewReader(acceptorConn).ReadString('\n')
	scoutAcceptorChannel <- response
}

//Commander talks to one acceptor; tells leader the acceptor's response
func (self *Leader) CommTalkToAcceptor(processPort string, commAcceptorChannel chan string, slotNum int, proposal string) {
	acceptorConn, err := net.Dial("tcp", "127.0.0.1:"+processPort)
	if err != nil {
		fmt.Println(err)
		commAcceptorChannel <- ""
		return
	}
	fmt.Fprintf(acceptorConn, "p2a,"+self.pid+","+strconv.Itoa(self.ballotNum)+" "+strconv.Itoa(slotNum)+" "+proposal+"\n")
	response, _ := bufio.NewReader(acceptorConn).ReadString('\n')
	commAcceptorChannel <- response
}



