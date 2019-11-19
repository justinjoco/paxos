package main

import (
	"os"
	"strconv"
)

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

	Leader := Leader{pid: id, ballotNum: 0, replicas: replicas, acceptors: acceptors} //ballot starts at zero - our choice
	Acceptor := Acceptor{pid: id, leaderFacingPort: leaderFacingPort}
	Replica := Replica{pid: id, masterFacingPort: masterFacingPort, commanderFacingPort: commanderFacingPort,
		chatLog: make(map[int]string)}

	go Leader.Run(replicaLeaderChannel)
	go Acceptor.Run()
	Replica.Run(replicaLeaderChannel) //Replica runs on main; others are parallel go routines (threads)

	os.Exit(0)

}
