package main

import (
	"os"
	"strconv"
)


func main(){

	args :=os.Args[1:4]


	id := args[0]
	n, _ := strconv.Atoi(args[1])
	masterFacingPort :=  args[2]
	id_num, _ := strconv.Atoi(id)

	replicaLeaderChannel := make(chan string)
	peerFacingPort := strconv.Itoa(20000+ id_num)

	var peers []string


	for i:=0 ; i < n; i++ {
		peerStr := strconv.Itoa(20000+i)
		peers = append(peers, peerStr)
	}


	Leader:= Leader{pid: id, ballotNum: 0}		//ballot starts at zero - our choice
	Acceptor:= Acceptor{pid: id, peers: peers, peerFacingPort: peerFacingPort}
	Replica:= Replica{pid: id, masterFacingPort: masterFacingPort, chatLog: make(map[int]string)}


	go Leader.Run(replicaLeaderChannel)
	go Acceptor.Run()
	Replica.Run(replicaLeaderChannel)			//Replica runs on main; others are parallel go routines (threads)

	os.Exit(0)

}
