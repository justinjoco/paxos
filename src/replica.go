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
}

const (
	CONNECT_HOST = "localhost"
	CONNECT_TYPE = "tcp"
)

func (self *Replica) Propose(proposal string, replicaLeaderChannel chan string) {
	max_slot := 0
	for k, v := range self.decisions {
		if v == proposal {
			return
		}
		if k > max_slot {
			max_slot = k
		}
	}
	next_slot := max_slot + 1
	self.proposals[next_slot] = proposal

	msgToLeader := "propose " + strconv.Itoa(next_slot) + " " + proposal //send proposal to leader
	replicaLeaderChannel <- msgToLeader
}

func (self *Replica) Perform(proposal string, connMaster net.Conn) {
	incremented := false
	for s, p := range self.decisions {
		if p == proposal && s < self.slot {
			self.slot += 1
			incremented = true
		}
	}
	if !incremented {

		// send back to master
		// split the proposal into msgID and actual msg
		proposalSlice := strings.Split(proposal, " ")
		msg := proposalSlice[1]
		msgID := proposalSlice[0]

		self.chatLog[self.slot] = msg
		self.slot += 1

		retMessage := "ack "
		retMessage += msgID + " "
		retMessage += strconv.Itoa(self.slot)
		lenStr := strconv.Itoa(len(retMessage))
		retMessage = lenStr + "-" + retMessage
		connMaster.Write([]byte(retMessage))
	}

}

func (self *Replica) Run(replicaLeaderChannel chan string) {

	lMaster, error := net.Listen(CONNECT_TYPE, CONNECT_HOST+":"+self.masterFacingPort)
	lCommander, error := net.Listen(CONNECT_TYPE, CONNECT_HOST+":"+self.commanderFacingPort)

	// Call acceptor.run() to start the acceptor process
	// Call leader.run() to start the leader process
	// this leader, acceptor, and replica share memory and talk to each other
	// via channels, not ports.

	// replica, leader, scout, commander, will communicate via
	// channels.
	// scout and commander will talk to acceptors via port.
	if error != nil {
		fmt.Println("Error listening!")
	}
	connMaster, error := lMaster.Accept()
	go self.HandleCommander(lCommander, connMaster, replicaLeaderChannel) // TO listen to decisions by other process's commanders

	self.HandleMaster(connMaster, replicaLeaderChannel)

}

func (self *Replica) HandleCommander(lCommander net.Listener, connMaster net.Conn, replicaLeaderChannel chan string) {
	defer lCommander.Close()
	//msg := ""
	connCommander, error := lCommander.Accept()
	reader := bufio.NewReader(connCommander)
	for {

		if error != nil {
			fmt.Println("Error while accepting connection")
			continue
		}

		message, _ := reader.ReadString('\n')

		message = strings.TrimSuffix(message, "\n")
		messageSlice := strings.Split(message, " ")
		keyWord := messageSlice[0]

		//removeComma := 0
		retMessage := ""
		switch keyWord {
		case "decision":
			slotNum := messageSlice[1] // s
			command := messageSlice[2] // p
			slotInt, _ := strconv.Atoi(slotNum)
			self.decisions[slotInt] = command
			for {
				pprime := self.decisions[self.slot]
				if pprime == "" {
					break
				}
				pdprime := self.proposals[self.slot]
				if pdprime != "" && pdprime != pprime {
					self.Propose(pdprime, replicaLeaderChannel)
				}
				self.Perform(pprime, connMaster)
			}

		default:

			retMessage += "Invalid keyword, mustbe decision"
			connCommander.Write([]byte(retMessage))

		}

	}

	//connMaster.Close()

}

func (self *Replica) HandleMaster(connMaster net.Conn, replicaLeaderChannel chan string) {
	msgId := ""
	msg := ""
	reader := bufio.NewReader(connMaster)
	for {
		/*
			if error != nil {
				fmt.Println("Error while accepting connection")
				continue
			}*/

		request, _ := reader.ReadString('\n')

		request = strings.TrimSuffix(request, "\n")
		requestSlice := strings.Split(request, " ")
		command := requestSlice[0]

		retMessage := ""
		//removeComma := 0
		switch command {
		case "msg":
			msgId = requestSlice[1]
			msg = requestSlice[2]
			self.Propose(msgId+" "+msg, replicaLeaderChannel)

		case "get":
			retMessage += "chatLog "
			// iterate through the chatlog
			//		msgCount := 0
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
			// TODO: messages for the master
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

func (self *Replica) ReceivePeers(lPeer net.Listener) {
	defer lPeer.Close()

	for {
		connPeer, error := lPeer.Accept()

		if error != nil {
			fmt.Println("Error while accepting connection")
			continue
		}

		message, _ := bufio.NewReader(connPeer).ReadString('\n')
		message = strings.TrimSuffix(message, "\n")
		if message == "ping" {
			connPeer.Write([]byte(self.pid))
		}
		connPeer.Close()

	}

}
