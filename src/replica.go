package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Replica struct {
	pid                 string
	masterFacingPort    string
	commanderFacingPort string
	chatLog             map[int]string
	proposals           map[int]string
	decisions           map[int]string //the key is the slot number
	slot                int
	n                   int
}

const (
	CONNECT_HOST = "localhost"
	CONNECT_TYPE = "tcp"
)

//Wrapper that runs replica
func (self *Replica) Run(replicaLeaderChannel chan string) {

	lMaster, error := net.Listen(CONNECT_TYPE, CONNECT_HOST+":"+self.masterFacingPort)
	lCommander, error := net.Listen(CONNECT_TYPE, CONNECT_HOST+":"+self.commanderFacingPort)

	if error != nil {
		fmt.Println("Error listening!")
	}
	connMaster, error := lMaster.Accept()
	

	go self.HandleCommander(lCommander, connMaster, replicaLeaderChannel) // TO listen to decisions by other process's commanders
	self.SyncDecisions(replicaLeaderChannel)
	self.HandleMaster(connMaster, replicaLeaderChannel)

}

//Ping other servers for their decision sets; enacted via leader
func (self *Replica) SyncDecisions(replicaLeaderChannel chan string){
	replicaLeaderChannel <- "catchup"
	numResponse := 0
	for numResponse < self.n{
	select {
		case response := <- replicaLeaderChannel:
			fmt.Println("GOT A RESPONSE OR NOT")
			if response != "" {
				responseSlice := strings.Split(response, ",")
				
				for _, decision := range responseSlice {
					decisionSlice := strings.Split(decision, " ")
					slot, _ := strconv.Atoi(decisionSlice[0])
					msg := decisionSlice[1]
					if self.decisions[slot] == ""{
						self.decisions[slot] = decision
						self.chatLog[slot] = msg
					} 
				}

			}
			numResponse +=1
		}
	}

	
}

//Propose a given command to all other servers
func (self *Replica) Propose(proposal string, replicaLeaderChannel chan string) {
	fmt.Println("PROPOSE: " + proposal)
	nextSlot := self.slot
	self.SyncDecisions(replicaLeaderChannel)

	fmt.Println("AFTER SYNCING DECISIONS")
	fmt.Println(self.decisions)
	//max := 0
	if self.decisions[nextSlot] != "" {
		for k, _ := range self.decisions {
			if nextSlot < k{
				nextSlot = k
			}

		}
		nextSlot += 1
		self.slot = nextSlot
	}
	
	self.proposals[nextSlot] = proposal
	fmt.Println("NEXT SLOT:" + strconv.Itoa(nextSlot))
	msgToLeader := "propose " + strconv.Itoa(nextSlot) + " " + proposal //send proposal to leader
	replicaLeaderChannel <- msgToLeader
}

//Applies everything in decisions onto chatlog
func (self *Replica) Perform(proposal string, connMaster net.Conn) {
	
	
	for s, p := range self.decisions {
		
		pSlice := strings.Split(p, " ")
		self.chatLog[s] = pSlice[1]
		
	}

	proposalSlice := strings.Split(proposal, " ")
	msgID := proposalSlice[0]

	retMessage := "ack "
	retMessage += msgID + " "
	retMessage += strconv.Itoa(self.slot)
	lenStr := strconv.Itoa(len(retMessage))
	retMessage = lenStr + "-" + retMessage

	connMaster.Write([]byte(retMessage))


}

//Replica responds to commander decision or sync messages
func (self *Replica) HandleCommander(lCommander net.Listener, connMaster net.Conn, replicaLeaderChannel chan string) {
	defer lCommander.Close()

	for {
		connCommander, error := lCommander.Accept()
	
		if error != nil {
			fmt.Println("Error while accepting connection")
			continue
		}
		reader := bufio.NewReader(connCommander)
		message, _ := reader.ReadString('\n')
		fmt.Println(message)
		message = strings.TrimSuffix(message, "\n")
		messageSlice := strings.Split(message, " ")
		keyWord := messageSlice[0]

		retMessage := ""
		switch keyWord {
		case "decision":
			fmt.Println(message)
			slotNum := messageSlice[1] // s
			command := strings.Join(messageSlice[2:], " ") // p
			slotInt, _ := strconv.Atoi(slotNum)
			
			fmt.Println(self.pid + " RECEIVED DECISION")
			self.decisions[slotInt] = command
			fmt.Println(self.decisions)

			self.Perform(command, connMaster)
		case "catchup":
			fmt.Println("RECEIVED CATCHUP")
			decisionSlice := make([]string, 0)
			retMessage := ""
			if len(self.decisions) > 0{ 
				for _, decision := range self.decisions {
					decisionSlice = append(decisionSlice, decision)
				}
				retMessage += strings.Join(decisionSlice,",")
			}
			fmt.Println("CATCHUP RETURN MSG:" + retMessage)
			connCommander.Write([]byte(retMessage + "\n"))

		default:

			retMessage += "Invalid keyword, mustbe decision"
			connCommander.Write([]byte(retMessage+ "\n") )

		}
		connCommander.Close()
	}


}

//Handle master commands; run Paxos on msg requests, crashes when told to do so
func (self *Replica) HandleMaster(connMaster net.Conn, replicaLeaderChannel chan string) {
	msgId := ""
	msg := ""
	reader := bufio.NewReader(connMaster)

	for {
	

		request, _ := reader.ReadString('\n')

		request = strings.TrimSuffix(request, "\n")
		requestSlice := strings.Split(request, " ")
		command := requestSlice[0]
		fmt.Println("PID:"+ self.pid + " " + request)
		retMessage := ""

		switch command {
			case "msg":
				msgId = requestSlice[1]
				self.slot, _ = strconv.Atoi(msgId)
				fmt.Println("SLOT: " + strconv.Itoa(self.slot))
				msg = requestSlice[2]
				found := false
				foundSlot := -1
				
				chatLog := self.chatLog
				for slot, savedMsg := range chatLog{
					if savedMsg == msg {
						foundSlot = slot
						found = true
						break
					}
				}

				if !found {
					self.Propose(msgId+" "+msg, replicaLeaderChannel)
				} else{
					retMessage := "ack "
					retMessage += msgId + " "
					retMessage += strconv.Itoa(foundSlot)
					lenStr := strconv.Itoa(len(retMessage))
					retMessage = lenStr + "-" + retMessage
					connMaster.Write([]byte(retMessage))	
				}
					
				
			case "get":
				retMessage += "chatLog "
				fmt.Println(self.pid + " CHATLOG")
				fmt.Println(self.chatLog)
				removeComma := 0
				counter := 0
				for i := 0; i <= 100; i++ {
					if counter == len(self.chatLog) {
						break
					}
					if self.chatLog[i] != "" {
						retMessage += self.chatLog[i] + ","
						removeComma = 1
						counter += 1
					}
				}
				retMessage = retMessage[0 : len(retMessage)-removeComma]
				lenStr := strconv.Itoa(len(retMessage))
				retMessage = lenStr + "-" + retMessage
				connMaster.Write([]byte(retMessage))

			case "crash":
				os.Exit(1)

			case "crashAfterP1b":
				crashStage = "p1b"
			case "crashAfterP2b":
				crashStage = "p2b"
			case "crashP1a":
				crashStage = "p1a"
				crashAfterSentTo = requestSlice[1:]
			case "crashP2a":
				crashStage = "p2a"
				crashAfterSentTo = requestSlice[1:]

			case "crashDecision":
				crashStage = "decision"
				crashAfterSentTo = requestSlice[1:]

			default:
				retMessage += "Invalid command. Use 'get', 'alive', or 'broadcast <message>'"
				connMaster.Write([]byte(retMessage))
		}

	}

	connMaster.Close()

}


