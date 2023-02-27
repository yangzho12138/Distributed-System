package main

import (
	"bufio"
	"container/heap"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	TimeStamp     string // ID of a message
	DeliverStatus bool   // true-delivered, false-not delivered
	Priority      int
	Sender        int    // which node sends the message
	Transaction   string // content of transaction
}

type Node struct {
	Address string
	Port    string
}

type PP struct {
	Sender   int
	Priority int
}

// msg json format
type MsgJson struct {
	MsgType   string `json:"msgType"`
	Sender    string `json:"sender"`
	TimeStamp string `json:"timestamp"`
	Content   string `json:"content"`
}

// collect all the proposed priority and find the max priority and its sender
var Pp map[string][]PP

var Account map[string]int // user account

var NodesToPorts map[string]Node // the port number different nodes listen to (read from config file)

var PortsToNodes map[string]string

var DialConnections map[string]net.Conn // store the connections

// read from command line
var node string

var configFilePath string

// priority for a process
var proposedPriority int

// priority to store the multicasts
var pq PriorityQueue

// read from config file
var nodeNum int

// heap interface -> priority queue
type PriorityQueue []Message

// methods of PriorityQueue
func (pq PriorityQueue) Len() int {
	return len(pq)
}
func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].Priority == pq[j].Priority { // break ties
		return pq[i].Sender > pq[j].Sender
	}
	return pq[i].Priority < pq[j].Priority
}
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}
func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(Message))
}
func (pq PriorityQueue) Update(timestamp string, priority int, sender int) {
	for i := 0; i < len(pq); i++ {
		if pq[i].DeliverStatus == false && pq[i].TimeStamp == timestamp {
			pq[i].Priority = priority
			pq[i].DeliverStatus = true
			pq[i].Sender = sender
		}
	}
}
func (pq *PriorityQueue) Top() interface{} {
	old := *pq
	n := len(old)
	x := old[n-1]
	return x
}

func ReadFile(path string) {
	NodesToPorts = make(map[string]Node)
	PortsToNodes = make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Read files failed")
	}

	buf := bufio.NewReader(f)
	line, err := buf.ReadString('\n')
	if err != nil {
		log.Fatal("Config file structure is uncorrect")
	}

	nodes, _ := strconv.Atoi(strings.TrimSpace(line))
	nodeNum = nodes

	for d := 0; d < nodeNum; d++ {
		line, _ := buf.ReadString('\n')
		nodeInfo := strings.Split(line, " ")
		n := Node{strings.TrimSpace(nodeInfo[1]), strings.TrimSpace(nodeInfo[2])}
		NodesToPorts[nodeInfo[0]] = n
		PortsToNodes[strings.TrimSpace(nodeInfo[1])] = nodeInfo[0]
		fmt.Println("Nodes to Ports ", NodesToPorts)
		fmt.Println("Ports to Nodes ", PortsToNodes)
	}
}

func ProcessTransaction(msg Message) {
	t := msg.Transaction
	tInfo := strings.Split(t, " ")
	if tInfo[0] == "DEPOSIT" {
		user := tInfo[1]
		amount, _ := strconv.Atoi(tInfo[2])
		Account[user] = Account[user] + amount
	} else if tInfo[0] == "TRANSFER" {
		userFrom := tInfo[1]
		userTo := tInfo[3]
		amount, _ := strconv.Atoi(tInfo[4])
		if Account[userFrom] > amount {
			Account[userFrom] = Account[userFrom] - amount
			Account[userTo] = Account[userTo] + amount
		}
	}

	fmt.Print("BALANCES ")
	for key, value := range Account {
		if value != 0 {
			fmt.Print(key + ":" + strconv.Itoa(value) + " ")
		}
	}
	fmt.Print("\n")

}

// deal with deliverable msg in pq
func ProcessPQ() {
	for {
		top, _ := pq.Top().(Message)
		if top.DeliverStatus == false {
			break
		}
		msg, _ := pq.Pop().(Message)
		ProcessTransaction(msg)
	}
}

func receiveMsg(conn net.Conn) {
	defer conn.Close()

	for {
		var msgJson MsgJson
		err := json.NewDecoder(conn).Decode(&msgJson)
		if err != nil {
			log.Println(err)
			return
		}

		msgType := msgJson.MsgType
		sender, _ := strconv.Atoi(msgJson.Sender)
		timestamp := msgJson.TimeStamp
		content := msgJson.Content

		if msgType == "T" {
			// store msg in pq
			p := proposedPriority
			proposedPriority++
			message := Message{timestamp, false, p, sender, content}
			pq.Push(message)
			// send proposed priority
			go sendMsg(strconv.Itoa(p), "PP", timestamp)
		} else if msgType == "PP" {
			p, _ := strconv.Atoi(content)
			Pp[timestamp] = append(Pp[timestamp], PP{sender, p})

			maxP := 0
			var maxPSender int
			if len(Pp[timestamp]) == nodeNum {
				for n := 0; n < nodeNum; n++ {
					if maxP < Pp[timestamp][n].Priority {
						maxP = Pp[timestamp][n].Priority
						maxPSender = Pp[timestamp][n].Sender
					}
				}
			}
			// update own msg in pq
			pq.Update(timestamp, maxP, maxPSender)
			// send agreed priority
			go sendMsg(strconv.Itoa(maxP)+"|"+strconv.Itoa(maxPSender), "PA", timestamp)

			// deal with deliverable msg in pq
			ProcessPQ()
		} else if msgType == "PA" {
			pa := strings.Split(content, "|")
			maxP, _ := strconv.Atoi(pa[0])
			maxPSender, _ := strconv.Atoi(pa[1])

			// update own msg in pq
			pq.Update(timestamp, maxP, maxPSender)

			// deal with deliverable msg in pq
			ProcessPQ()
		}
	}
}

func Multicast(msg string, msgType string, timestamp string) {
	// multicast
	for key, _ := range DialConnections {
		// fmt.Println(key + " " + value.Address + " " + value.Port)
		if key != node {
			// send transaction msg to other nodes
			conn := DialConnections[key]
			// json the msg
			msgJson := MsgJson{MsgType: msgType, Sender: node[4:], TimeStamp: timestamp, Content: msg}
			err := json.NewEncoder(conn).Encode(msgJson)
			if err != nil {
				fmt.Println("Error encoding JSON:", err)
				return
			}
		}
	}
}

// send msg format: Type + Sender + Timestamp + Content
// Type: T(Transaction) / PP(Priority Proposed) / PA(Priority Agreed)
// Content: T-Transaction Content; PP-Proposed Priority; PA-Agreed Priority|Agreed Priority Sender
func sendMsg(msg string, msgType string, timestamp string) {
	if msgType == "T" {
		// create msg and store in pq
		sender, _ := strconv.Atoi(node[4:])
		p := proposedPriority
		proposedPriority++
		message := Message{timestamp, false, p, sender, msg}
		Pp[timestamp] = append(Pp[timestamp], PP{sender, p})
		pq.Push(message)

		// multicast
		Multicast(msg, msgType, timestamp)
	} else if msgType == "PP" {
		Multicast(msg, msgType, timestamp)
	} else if msgType == "PA" {
		Multicast(msg, msgType, timestamp)
	}
}

// send transaction msg generated by scripts
func send() {
	// send
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
		go sendMsg(s.Text(), "T", timestamp)
	}
}

func monitor() {
	for {
		for i := 1; i <= nodeNum; i++ {
			n := "node" + strconv.Itoa(i)
			value := NodesToPorts[n]
			conn, ok := DialConnections[n]

			if ok {
				continue
			}
			conn, err := net.Dial("tcp", value.Address+":"+value.Port)
			if err != nil {
				continue
			}
			DialConnections[n] = conn
			fmt.Println("Successfully established connection with  ", n)
			defer conn.Close()
		}
	}
}

func initialize() {
	// init proposed priority
	proposedPriority = 1
	// initial map
	Account = make(map[string]int)
	Pp = make(map[string][]PP)
	DialConnections = make(map[string]net.Conn)
	pq := &PriorityQueue{}
	heap.Init(pq)
}

func main() {
	if len(os.Args) > 1 {
		node = os.Args[1]
		configFilePath = os.Args[2]
	} else {
		log.Println("Please enter the node number and config file in the command line")
	}

	ReadFile(configFilePath)

	// listen to port
	listener, err := net.Listen("tcp", ":"+NodesToPorts[node].Port)
	if err != nil {
		log.Println("Failed to listen on port ", err)
	}
	defer listener.Close()

	fmt.Println("Listen successfully")

	time.Sleep(10e9)

	initialize()

	go monitor()

	// send message
	go send()

	for {
		// listen
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Failed to receive message ", err)
			return
		}

		// go func(conn net.Conn) {
		// 	input := bufio.NewReader(conn)
		// 	pattern, err := input.ReadString('\n')
		// 	fmt.Println("input is ", input)
		// 	fmt.Println("pattern is ", pattern)
		// 	fmt.Println("error is ", err)
		// 	// remove connection from DialConnections
		// }(conn)

		go receiveMsg(conn)
	}
}
