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
	chatLog          []string
}

const (
	CONNECT_HOST = "localhost"
	CONNECT_TYPE = "tcp"
)

func (self *Replica) Run() {

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

	connMaster, error := lMaster.Accept()
	reader := bufio.NewReader(connMaster)
	for {

		if error != nil {
			fmt.Println("Error while accepting connection")
			continue
		}

		message, _ := reader.ReadString('\n')

		message = strings.TrimSuffix(message, "\n")
		messageSlice := strings.Split(message, " ")
		command := messageSlice[0]

		retMessage := ""
		removeComma := 0
		switch command {
		case "alive":
			retMessage += "alive "
			for _, port := range self.alive {
				retMessage += port + ","
				removeComma = 1
			}
			retMessage = retMessage[0 : len(retMessage)-removeComma]
			lenStr := strconv.Itoa(len(retMessage))
			retMessage = lenStr + "-" + retMessage

		case "get":
			retMessage += "messages "
			for _, message := range self.messages {
				retMessage += message + ","
				removeComma = 1
			}
			retMessage = retMessage[0 : len(retMessage)-removeComma]
			lenStr := strconv.Itoa(len(retMessage))
			retMessage = lenStr + "-" + retMessage

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
