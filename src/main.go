package main

import (
//	"bufio"
//	"net"
	"os"
	"strconv"
//	"fmt"
//	"time"
)

var crashStage string
var crashAfterSentTo []string

/*

 func Heartbeat(pid string, leaderFacingPort string, n int, acceptors []string) { //maintain alive list; calculate the majority

 //	fmt.Println(acceptors)
 	tempAlive := make([]string, 0)
 	for {

 		tempAlive = tempAlive[:0]

 		for _, otherPort := range acceptors {
 			//fmt.Println(otherPort)
 				if otherPort != leaderFacingPort {
	 				acceptorConn, err := net.Dial("tcp", "127.0.0.1:"+otherPort)
	 				if err != nil {
	 					fmt.Println(err)
	 					continue
					}

	 				fmt.Fprintf(acceptorConn, "ping\n")
	 				reader := bufio.NewReader(acceptorConn)
	 				response, _ := reader.ReadString('\n')
	 		//		fmt.Println(response)
	 				tempAlive = append(tempAlive, response)
	 				acceptorConn.Close()
 				}
			}
		tempAlive = append(tempAlive, pid)
 		fmt.Println("PID:" + pid)
 		fmt.Println(tempAlive)
 		if len(tempAlive) == n {
 			break
 		} 

 		time.Sleep(500 * time.Millisecond)
 	}
 }

*/

func main() {

	args := os.Args[1:4]

	pid := args[0]
	n, _ := strconv.Atoi(args[1])
	masterFacingPort := args[2]
	id_num, _ := strconv.Atoi(pid)

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

	leader := Leader{pid: pid, ballotNum: 0, replicas: replicas, acceptors: acceptors, proposals: make(map[int]string)} //ballot starts at zero - our choice
	acceptor := Acceptor{pid: pid, leaderFacingPort: leaderFacingPort}
	replica := Replica{pid: pid, masterFacingPort: masterFacingPort, commanderFacingPort: commanderFacingPort,
		chatLog: make(map[int]string), proposals: make(map[int]string), decisions: make(map[int]string)}

	go acceptor.Run()
	go replica.Run(replicaLeaderChannel) //Leader runs on main; others are parallel go routines (threads)
	
//	Heartbeat(pid, leaderFacingPort, n, acceptors)
//	fmt.Println("ALL ALIVE")
	leader.Run(replicaLeaderChannel)
	

	os.Exit(0)

}
