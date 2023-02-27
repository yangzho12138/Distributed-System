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

type Transaction struct {
	TransactionId     string // ID of a transaction
	DeliverStatus bool   // true-delivered, false-not delivered
	Priority      int
	Sender        int    // which node sends the message
	Content   string // content of transaction
}

type Node struct {
	Address string
	Port    string
}

type SequenceObject struct {
	Sender   int
	Priority int
}

// msg json format
type MsgJson struct {
	MsgType   string `json:"msgType"`
	Sender    string `json:"sender"`
	TransactionId string `json:"id"`
	Content   string `json:"content"`
}

// store a list of transaction and their proposed priorities by all the sender
var SequenceOrdering map[string][]SequenceObject

// bank accounts with balance
var Account map[string]int 

// <node, address and port on which the node is running>
var NodesToPorts map[string]Node 

var PortsToNodes map[string]string

// store the connections 
var DialConnections map[string]net.Conn 

// node running on the host server
var hostNode string

// file path with config information
var configFilePath string

// next available priority value to be proposed
var currentPriority int

// priority queue to store the transactions 
var pq PriorityQueue

// total number of nodes in the cluster
var nodeNum int

// heap interface -> priority queue
type PriorityQueue []Transaction

// methods for pq
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
	*pq = append(*pq, x.(Transaction))
}
func (pq PriorityQueue) Update(transactionId string, priority int, sender int, msgType string) {
	for i := 0; i < len(pq); i++ {
		if pq[i].DeliverStatus == false && pq[i].TransactionId == transactionId {
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
		fmt.Println("Ports to Nodes ", PortsToNodes) // TODO: not sure if we need this
	}
}

func ProcessTransaction(transaction Transaction) {
	content := transaction.Content
	transInfo := strings.Split(content, " ")
	if transInfo[0] == "DEPOSIT" {
		user := transInfo[1]
		amount, _ := strconv.Atoi(transInfo[2])
		Account[user] = Account[user] + amount
	} else if transInfo[0] == "TRANSFER" {
		userFrom := transInfo[1]
		userTo := transInfo[3]
		amount, _ := strconv.Atoi(transInfo[4])
		if Account[userFrom] >= amount { // TODO: do we need some sort of mutex here
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

// deliver a transaction from the front of pq
func ProcessPQ() {
	for {
		top, _ := pq.Top().(Transaction)
		if top.DeliverStatus == false {
			break
		}
		transaction, _ := pq.Pop().(Transaction)
		ProcessTransaction(transaction)
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

		content := msgJson.Content
		msgType := msgJson.MsgType
		transactionId := msgJson.TransactionId
		sender, _ := strconv.Atoi(msgJson.Sender)

		if msgType == "T" {
			// received message strcture: <transaction content, "T", transaction id, sender>
			proposedPriority := currentPriority
			currentPriority++
			SequenceOrdering[transactionId] = append(SequenceOrdering[transactionId], SequenceObject{sender, proposedPriority})
			transaction := Transaction{transactionId, false, proposedPriority, sender, content}
			pq.Push(transaction)

			// sent message structure: <proposed priority, "PP", transaction id>
			go sendMsg(strconv.Itoa(proposedPriority), "PP", transactionId)

		} else if msgType == "PP" {
			// received message structure: <proposed priority, "PP", transaction id, sender>
			proposedPriority, _ := strconv.Atoi(content)
			SequenceOrdering[transactionId] = append(SequenceOrdering[transactionId], SequenceObject{sender, proposedPriority})

			maxPriority := 0
			var maxPrioritySender int
			if len(SequenceOrdering[transactionId]) == nodeNum {
				for n := 0; n < nodeNum; n++ {
					if maxPriority < SequenceOrdering[transactionId][n].Priority {
						maxPriority = SequenceOrdering[transactionId][n].Priority
						maxPrioritySender = SequenceOrdering[transactionId][n].Sender
						// TODO: update SequenceOrdering ?
					}
				}
			}
			pq.Update(transactionId, maxPriority, maxPrioritySender, msgType)

			// sent message structure: <agreed priority | agreed priority sender, "PA", transaction id>
			go sendMsg(strconv.Itoa(maxPriority)+"|"+strconv.Itoa(maxPrioritySender), "PA", transactionId)

			ProcessPQ() // TODO: shouldn't mark the transaction as deliverable at this point

		} else if msgType == "PA" {
			// received message structure: <agreed priority | agreed priority sender, "PA", transaction id, sender>
			agreedPriorityInfo := strings.Split(content, "|")
			maxPriority, _ := strconv.Atoi(agreedPriorityInfo[0])
			maxPrioritySender, _ := strconv.Atoi(agreedPriorityInfo[1])

			pq.Update(transactionId, maxPriority, maxPrioritySender, msgType)

			// process transaction
			ProcessPQ()
		}
	}
}

func Multicast(msg string, msgType string, transactionId string) {
	for key, _ := range DialConnections {
		if key != hostNode {	
			// send transaction msg to other nodes
			conn := DialConnections[key]

			// json the msg
			msgJson := MsgJson{Content: msg, MsgType: msgType, TransactionId: transactionId, Sender: hostNode[4:]}
			err := json.NewEncoder(conn).Encode(msgJson)
			if err != nil {
				fmt.Println("Error encoding JSON:", err)
				return
			}
		}
	}
}

func sendMsg(msg string, msgType string, transactionId string) {
	if msgType == "T" {
		// create the transaction and store it into pq
		sender, _ := strconv.Atoi(hostNode[4:])
		proposedPriority := currentPriority
		currentPriority++
		transaction := Transaction{transactionId, false, proposedPriority, sender, msg}
		SequenceOrdering[transactionId] = append(SequenceOrdering[transactionId], SequenceObject{sender, proposedPriority})
		pq.Push(transaction)

		// multicast the new transaction to other nodes
		// <transaction content, "T", transaction id>
		Multicast(msg, msgType, transactionId)

	} else if msgType == "PP" {
		// multicast the proposed priority to other nodes
		// <proposed priority, "PP", transaction id>
		Multicast(msg, msgType, transactionId)

	} else if msgType == "PA" {
		// multicast the agreed priority to other nodes
		// <agreed priority | agreed priority sender, "PP", transaction id>
		Multicast(msg, msgType, transactionId)
	}
}

// send a new transaction generated by the script
func sendTransaction() {
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
	// current priority is 1 at the beginning
	currentPriority = 1

	// initialize
	Account = make(map[string]int)
	SequenceOrdering = make(map[string][]SequenceObject)
	DialConnections = make(map[string]net.Conn)
	pq := &PriorityQueue{}
	heap.Init(pq)
}

func main() {
	if len(os.Args) > 1 {
		hostNode = os.Args[1]
		configFilePath = os.Args[2]
	} else {
		log.Println("Please enter the node number and config file in the command line")
	}

	ReadFile(configFilePath)

	// listen on port
	listener, err := net.Listen("tcp", ":"+NodesToPorts[hostNode].Port)
	if err != nil {
		log.Println("Failed to listen on port ", err)
	}
	defer listener.Close()

	fmt.Println("Listen successfully")

	time.Sleep(10e9)

	initialize()

	go monitor()

	// send a new transaction
	go sendTransaction()

	for {
		// listen to other nodes
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Failed to receive message ", err)
			return
		}

		// TODO: I think we should handle disconnected nodes here
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
