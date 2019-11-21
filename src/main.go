package main

import (
	"fmt"
	"os"
	"strconv"
)

var crashStage string
var crashAfterSentTo []string

func main() {

	args := os.Args[1:4]

	pid := args[0]
	n, _ := strconv.Atoi(args[1])
	masterFacingPort := args[2]
	id_num, _ := strconv.Atoi(pid)

	replicaLeaderChannel := make(chan string)
	leaderFacingPort := strconv.Itoa(20000 + id_num)
	commanderFacingPort := strconv.Itoa(25000 + id_num)

	acceptors := make(map[int]string)
	replicas := make(map[int]string)

	for i := 0; i < n; i++ {
		acceptorStr := strconv.Itoa(20000 + i)
		acceptors[i] = acceptorStr
		replicaStr := strconv.Itoa(25000 + i)
		replicas[i] = replicaStr
	}

	//Set up leader, acceptor, replica structs
	leader := Leader{pid: pid, ballotNum: 0, replicas: replicas, acceptors: acceptors, proposals: make(map[int]string)} //ballot starts at zero - our choice
	acceptor := Acceptor{pid: pid, leaderFacingPort: leaderFacingPort, currentBallot: -1, accepted: make(map[int]string)}
	replica := Replica{pid: pid, masterFacingPort: masterFacingPort, commanderFacingPort: commanderFacingPort,
		chatLog: make(map[int]string), proposals: make(map[int]string), decisions: make(map[int]string), n: n}

	//Run acceptor and replica on separate go routines; leader is on main goroutine.
	//Replica and leader talk to each other via channel
	fmt.Println("STARTING UP:: " + pid)
	go acceptor.Run()
	go leader.Run(replicaLeaderChannel)
	replica.Run(replicaLeaderChannel)

}
