package main

import (
	"bufio"
	"fmt"
	"net"
	"sort"
	"time"
)

type Acceptor struct {
	pid   string
	peers []string //might be the list of acceptor ports to talk to
	alive int      //should just be the count of alive processes for
	//majority calculation purposes
}

func (self *Acceptor) run() {

}

//Heartbeat should hhappen here

func (self *Acceptor) Heartbeat(broadcastMode bool, message string) {

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
