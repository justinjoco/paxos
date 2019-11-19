package main

import (
	"os"
	"strconv"
)

var crashStage string
var crashAfterSentTo []string

func main() {

	args := os.Args[1:4]

	id := args[0]
	n, _ := strconv.Atoi(args[1])
	masterFacingPort := args[2]
	id_num, _ := strconv.Atoi(id)

	replicaLeaderChannel := make(chan string)
	leaderFacingPort := strconv.Itoa(20000 + id_num)
	commanderFacingPort := strconv.Itoa(20100 + id_num)

	var acceptors []string
	var replicas []string

	for i := 0; i < n; i++ {
		acceptorStr := strconv.Itoa(20000 + i)
		acceptors = append(acceptors, acceptorStr)
		replicaStr := strconv.Itoa(20100 + i)
		replicas = append(replicas, replicaStr)
	}

	leader := Leader{pid: id, ballotNum: 0, replicas: replicas, acceptors: acceptors, proposals: make(map[int]string)} //ballot starts at zero - our choice
	acceptor := Acceptor{pid: id, leaderFacingPort: leaderFacingPort}
	replica := Replica{pid: id, masterFacingPort: masterFacingPort, commanderFacingPort: commanderFacingPort,
		chatLog: make(map[int]string), proposals: make(map[int]string), decisions: make(map[int]string)}

	go leader.Run(replicaLeaderChannel)
	go acceptor.Run()
	replica.Run(replicaLeaderChannel) //Replica runs on main; others are parallel go routines (threads)

	os.Exit(0)

}
