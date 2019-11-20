package main

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
	"fmt"
)

type Acceptor struct {
	pid              string
	leaderFacingPort string
	currentBallot    int
	accepted         []string
}

func (self *Acceptor) Run() {
	fmt.Println("LEADER FACING PORT: " + self.leaderFacingPort)
	lLeader, _ := net.Listen(CONNECT_TYPE, CONNECT_HOST+":"+self.leaderFacingPort)

	defer lLeader.Close()
	//	msg := ""
	
//	if err != nil {
	//	fmt.Println("error while accepting connection")
		//continue
//	}
	//reader := bufio.NewReader(connLeader)
	for {
		connLeader, err := lLeader.Accept()
		if err != nil {
			fmt.Println("error while accepting connection")
			continue
		}
		reader := bufio.NewReader(connLeader)
		message, _ := reader.ReadString('\n')
		fmt.Println(message)
		message = strings.TrimSuffix(message, "\n")
		messageSlice := strings.Split(message, ",")
		keyWord := messageSlice[0]
		//fmt.Println(keyWord)
		retMessage := ""
		switch keyWord {
		case "p1a":
			//	leaderId := messageSlice[1]  // lambda
			receivedBallot := messageSlice[2] // b
			receivedBallotInt, _ := strconv.Atoi(receivedBallot)
			if receivedBallotInt > self.currentBallot {
				fmt.Println("ACCEPTED NEW BALLOT: " + receivedBallot)
				self.currentBallot = receivedBallotInt
			}
			retMessage += "p1b," + self.pid + "," + strconv.Itoa(self.currentBallot)
			acceptedStr := ""
			for _, accepted := range self.accepted {
				acceptedStr += "," + accepted
			}
			retMessage += acceptedStr
			connLeader.Write([]byte(retMessage + "\n"))
			if crashStage == "p1b" {
				os.Exit(1)
			}

		case "p2a":
			//	leaderId := messageSlice[1]  // lambda
			pval := messageSlice[2]
			fmt.Println(pval)
			pvalSlice := strings.Split(pval, " ")
			receivedBallotInt, _ := strconv.Atoi(pvalSlice[0])
			if receivedBallotInt >= self.currentBallot {
				self.currentBallot = receivedBallotInt
				self.accepted = append(self.accepted, pval)
			}
			retMessage += "p2b," + self.pid + "," + strconv.Itoa(self.currentBallot) + "\n"
			fmt.Println(retMessage)
			connLeader.Write([]byte(retMessage))
			if crashStage == "p2b" {
				os.Exit(1)
			}
		//case "ping":
			//fmt.Println("SEND BACK PID")
		//	connLeader.Write([]byte(self.pid))
		default:
			retMessage += "Invalid keyword, must be p1a or p2a or ping"
			connLeader.Write([]byte(retMessage+ "\n"))

		}
		connLeader.Close()
	}

}

//Heartbeat should hhappen here
