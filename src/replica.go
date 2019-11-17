package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Replica struct {
	pid              string
	masterFacingPort string
	chatLog          map[int]string
	proposals				 map[int]string
	decisions				 map[int]string		//the key is the slot number
	slot						 int
}

const (
	CONNECT_HOST = "localhost"
	CONNECT_TYPE = "tcp"
)

func (self *Replica) Propose(proposal string, replicaLeaderChannel chan string){
	max_slot := 0
	for k, v := range self.decisions{
		if v == proposal{
			return
		}
		if k > max_slot{
			max_slot = k
		}
	}
	next_slot := max_slot + 1
	self.proposals[next_slot] = proposal

	msgToLeader := "propose " + strconv.Itoa(next_slot) +" "+ proposal	//send proposal to leader
	replicaLeaderChannel <- msgToLeader
}

func (self *Replica) Perform(proposal string){

}


func (self *Replica) Run(replicaLeaderChannel chan string) {

	lMaster, error := net.Listen(CONNECT_TYPE, CONNECT_HOST+":"+self.masterFacingPort)

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

	self.HandleMaster(lMaster)

}

func (self *Replica) HandleMaster(lMaster net.Listener) {
	defer lMaster.Close()
	msgId := ""
	msg := ""
	connMaster, error := lMaster.Accept()
	reader := bufio.NewReader(connMaster)
	for {

		if error != nil {
			fmt.Println("Error while accepting connection")
			continue
		}

		request, _ := reader.ReadString('\n')

		request = strings.TrimSuffix(request, "\n")
		requestSlice := strings.Split(request, " ")
		command := requestSlice[0]

		retMessage := ""
		removeComma := 0
		switch command {
		case "msg":
			retMessage += "ack "
			msgId = requestSlice[1]
			msg = requestSlice[2]
			self.propose(msg)
			//retMessage = retMessage[0 : len(retMessage)-removeComma]
			//lenStr := strconv.Itoa(len(retMessage))
			//retMessage = lenStr + "-" + retMessage

		case "get":
			retMessage += "messages "
			for _, request := range self.messages {
				retMessage += request + ","
				removeComma = 1
			}
			retMessage = retMessage[0 : len(retMessage)-removeComma]
			lenStr := strconv.Itoa(len(retMessage))
			retMessage = lenStr + "-" + retMessage

		case "crash":
			os.exit(1)

		case "crashAfterP1b":
			os.exit(1)

		case "crashAfterP2b":
			os.exit(1)

		case "crashP1a":
			os.exit(1)

		case "crashP2a":
			os.exit(1)

		case "crashDecision":
			os.exit(1)

		default:
			broadcastMessage := messageSlice[1]
			if broadcastMessage != "" {
				self.messages = append(self.messages, broadcastMessage)
				self.Heartbeat(true, broadcastMessage)
			} else {
				retMessage += "Invalid command. Use 'get', 'alive', or 'broadcast <message>'"
			}
		}

		connMaster.Write([]byte(retMessage))

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
		} else {
			self.messages = append(self.messages, message)
		}
		connPeer.Close()

	}

}
