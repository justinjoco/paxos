package main

import (
	"bufio"
	"fmt"
	"net"
	"sort"
	"time"
)

type Acceptor struct {
	pid							string
	peers						[]string //might be the list of acceptor ports to talk to
	alive						int      //should just be the count of alive processes for
	peerFacingPort	string
	//majority calculation purposes
}

func (self *Acceptor) Run() {

}

//Heartbeat should hhappen here

func (self *Acceptor) Heartbeat(broadcastMode bool, message string) { //maintain alive list; calculate the majority

	for {

		tempAlive := make([]string, 0)

		for _, otherPort := range self.peers {

			if otherPort != self.peerFacingPort {
				peerConn, err := net.Dial("tcp", "127.0.0.1:"+otherPort)
				if err != nil {
					continue
				}

				fmt.Fprintf(peerConn, message+"\n")
				response, _ := bufio.NewReader(peerConn).ReadString('\n')
				tempAlive = append(tempAlive, response)
			}

		}

		if broadcastMode {
			break
		}

		tempAlive = append(tempAlive, self.pid)
		sort.Strings(tempAlive)
		self.alive = tempAlive
		time.Sleep(1000 * time.Millisecond)
	}
}
