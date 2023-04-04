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
	"sync"
	"sort"
)

type Transaction struct {
	TransactionId     string // ID of a transaction
	DeliverStatus bool   // true-delivered, false-not delivered
	Priority      int
	Sender        int    // which node sends the message
	Content   string // content of transaction
}

type Node struct {
	Id string
	Address string
	Port    string
	Connection net.Conn
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

// bank accounts with balance
// var Account map[string]int 
type Account struct{
	accountLock sync.RWMutex
	account map[string]int 
}

// store a list of transaction and their proposed priorities by all the sender
var SequenceOrdering map[string][]SequenceObject

var Accounts Account

// a list of node id to dns address/port mapping
var NodesToPorts map[string]Node 

// a list of ip address to node id mapping
var AddressToId map[string]string

// a list of actually joined nodes
var ConnectedNodes map[string]Node 

// lock for ConnectedNodes
var NodeLock sync.RWMutex

// lock for pq
var PQLock sync.RWMutex

// lock for SequenceOrdering
var SequenceLock sync.RWMutex

// node running on the host server
var hostNode Node

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
	return pq[i].Priority > pq[j].Priority
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
func (pq *PriorityQueue) Update(transactionId string, priority int, sender int, msgType string) {
	for i := range *pq {
		if (*pq)[i].DeliverStatus == false && (*pq)[i].TransactionId == transactionId {
	  		(*pq)[i].Priority = priority
	  		(*pq)[i].DeliverStatus = true
	  		(*pq)[i].Sender = sender
	  		heap.Fix(pq, i)
	  		return
	 	}
	}
}
   
func(pq *PriorityQueue) Top() interface{} {
	old := *pq
	n := len(old)
	if n == 0{
		return nil
	}
	x := old[n-1]
	return x
}

func ReadFile(path string) {

	NodesToPorts = make(map[string]Node)
	AddressToId = make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Read files failed")
	}

	buf := bufio.NewReader(f)
	line, err := buf.ReadString('\n')
	if err != nil {
		log.Fatal("Config file structure is incorrect")
	}

	nodes, _ := strconv.Atoi(strings.TrimSpace(line))
	nodeNum = nodes

	for d := 0; d < nodeNum; d++ {
		line, _ := buf.ReadString('\n')
		nodeInfo := strings.Split(line, " ")
		address, _ := net.LookupHost(strings.TrimSpace(nodeInfo[1]))
		
		node := Node {
			Id: strings.TrimSpace(nodeInfo[0]),
			Address: address[0], 
			Port: strings.TrimSpace(nodeInfo[2]),
		}
		NodesToPorts[node.Id] = node
		AddressToId[address[0]] = strings.TrimSpace(nodeInfo[0])
		if node.Id == os.Args[1] {
			hostNode = node
		}
	}
}

func ProcessTransaction(transaction Transaction) {
	content := transaction.Content
	transInfo := strings.Split(content, " ")
	if transInfo[0] == "DEPOSIT" {
		user := transInfo[1]
		amount, _ := strconv.Atoi(transInfo[2])
		Accounts.accountLock.Lock()
		Accounts.account[user] = Accounts.account[user] + amount
		Accounts.accountLock.Unlock()
	} else if transInfo[0] == "TRANSFER" {
		userFrom := transInfo[1]
		userTo := transInfo[3]
		amount, _ := strconv.Atoi(transInfo[4])
		Accounts.accountLock.Lock()
		if Accounts.account[userFrom] >= amount {
			Accounts.account[userFrom] = Accounts.account[userFrom] - amount
			Accounts.account[userTo] = Accounts.account[userTo] + amount
		}
		Accounts.accountLock.Unlock()
	}

	current := strconv.FormatInt(time.Now().UnixNano(), 10)
	
	f, _ := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f.WriteString(transaction.TransactionId + " " + current + "\n") // unit in nanosecond

	users := make([]string, 0, len(Accounts.account))
    for k := range Accounts.account {
        users = append(users, k)
    }
    sort.Strings(users)

	if len(users) > 0 {
		fmt.Print("BALANCES ")
		Accounts.accountLock.RLock()
		for _, key := range users {
			value := Accounts.account[key]
			if value != 0 {
				fmt.Print(key + ":" + strconv.Itoa(value) + " ")
			}
		}
		Accounts.accountLock.RUnlock()
		fmt.Print("\n")
	}

}

// deliver a transaction from the front of pq
func ProcessPQ(){
	for {
		t := pq.Top()
		if t == nil {
			break
		}
	 	top, _ := t.(Transaction)
	 	if top.DeliverStatus == false {
			break
		}
		msg, _ := pq.Pop().(Transaction)
		ProcessTransaction(msg)
	}
}
   
func receiveMsg(conn net.Conn, id string) {
	defer conn.Close()

	for {
		// deadline := time.Now().Add(5 * time.Second)
    	// conn.SetDeadline(deadline)
		var msgJson MsgJson
		err := json.NewDecoder(conn).Decode(&msgJson)
		if err != nil {

			if netErr, ok := err.(net.Error); ok && netErr.Timeout(){
				NodeLock.Lock()
				delete(ConnectedNodes, id)
				nodeNum--
				NodeLock.Unlock()
				return
			}
			continue
		}

		deadline := time.Now().Add(10 * time.Second)
    	conn.SetDeadline(deadline)

		content := msgJson.Content
		msgType := msgJson.MsgType
		transactionId := msgJson.TransactionId
		sender, _ := strconv.Atoi(string(msgJson.Sender[4]))

		if msgType == "T" {
			// received message strcture: <transaction content, "T", transaction id, sender>
			proposedPriority := currentPriority
			currentPriority++
			transaction := Transaction{transactionId, false, proposedPriority, sender, content}

			PQLock.Lock()
			pq.Push(transaction)
			PQLock.Unlock()

			// sent message structure: <proposed priority, "PP", transaction id>
			go sendMsg(strconv.Itoa(proposedPriority), "PP", transactionId, msgJson.Sender)

		} else if msgType == "PP" {
			// received message structure: <proposed priority, "PP", transaction id, sender>
			proposedPriority, _ := strconv.Atoi(content)

			SequenceLock.Lock()
			SequenceOrdering[transactionId] = append(SequenceOrdering[transactionId], SequenceObject{sender, proposedPriority})
			SequenceLock.Unlock()

			maxPriority := 0
			var maxPrioritySender int
			SequenceLock.RLock()
			if len(SequenceOrdering[transactionId]) == nodeNum {
				for n := 0; n < nodeNum; n++ {
					if maxPriority < SequenceOrdering[transactionId][n].Priority {
						maxPriority = SequenceOrdering[transactionId][n].Priority
						maxPrioritySender = SequenceOrdering[transactionId][n].Sender
					}
				}
				pq.Update(transactionId, maxPriority, maxPrioritySender, msgType)
				// sent message structure: <agreed priority | agreed priority sender, "PA", transaction id>
				go sendMsg(strconv.Itoa(maxPriority)+"|"+strconv.Itoa(maxPrioritySender), "PA", transactionId, msgJson.Sender)
				ProcessPQ()
			}
			SequenceLock.RUnlock()

		} else if msgType == "PA" {
			// received message structure: <agreed priority | agreed priority sender, "PA", transaction id, sender>
			agreedPriorityInfo := strings.Split(content, "|")
			maxPriority, _ := strconv.Atoi(agreedPriorityInfo[0])
			maxPrioritySender, _ := strconv.Atoi(agreedPriorityInfo[1])

			PQLock.Lock()
			pq.Update(transactionId, maxPriority, maxPrioritySender, msgType)
			PQLock.Unlock()

			// process transaction
			ProcessPQ()
		}
	}
}

func Multicast(msg string, msgType string, transactionId string) {
	NodeLock.RLock()
	for key, _ := range ConnectedNodes {
		if key != hostNode.Id {	
			// send transaction msg to other nodes
			if _, ok := ConnectedNodes[key]; ok{
				conn := ConnectedNodes[key].Connection

				// json the msg
				msgJson := MsgJson{Content: msg, MsgType: msgType, TransactionId: transactionId, Sender: hostNode.Id}
				json.NewEncoder(conn).Encode(msgJson)
				// if err != nil {
				// 	fmt.Println("Error encoding JSON:", err)
				// }
			}
		}
	}
	NodeLock.RUnlock()
}

func Unicast(msg string, msgType string, transactionId string, targetId string){
	NodeLock.RLock()
	if _, ok := ConnectedNodes[targetId]; ok{
		conn := ConnectedNodes[targetId].Connection

		msgJson := MsgJson{Content: msg, MsgType: msgType, TransactionId: transactionId, Sender: hostNode.Id}
		err := json.NewEncoder(conn).Encode(msgJson)
		if err != nil {
			fmt.Println("Error encoding JSON:", err)
		}
	}	
	NodeLock.RUnlock()
}

func sendMsg(msg string, msgType string, transactionId string, targetId string) {
	if msgType == "T" {
		// create the transaction and store it into pq
		sender, _ := strconv.Atoi(string(hostNode.Id[4]))
		proposedPriority := currentPriority
		currentPriority++
		transaction := Transaction{transactionId, false, proposedPriority, sender, msg}
		SequenceOrdering[transactionId] = append(SequenceOrdering[transactionId], SequenceObject{sender, proposedPriority})

		PQLock.Lock()
		pq.Push(transaction)
		PQLock.Unlock()

		// multicast the new transaction to other nodes
		// <transaction content, "T", transaction id>
		Multicast(msg, msgType, transactionId)

	} else if msgType == "PP" {
		// multicast the proposed priority to other nodes
		// <proposed priority, "PP", transaction id, >
		Unicast(msg, msgType, transactionId, targetId)

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
		go sendMsg(s.Text(), "T", timestamp, "none")
	}
}

func deleteFile(path string){
	e := os.Remove(path)
    if e != nil {
        log.Fatal(e)
    }
}

func removeFile(path string){
	if fileExists(path) {
		if err := os.Remove(path); err != nil {
			log.Fatal(err)
		}
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
	   return false
	}
	return !info.IsDir()
 }

func initialize() {
	ConnectedNodes = make(map[string]Node)
	// current priority is 1 at the beginning
	currentPriority = 1

	// initialize
	Accounts = Account {
		accountLock: sync.RWMutex{},
		account: make(map[string]int),
	}
	SequenceOrdering = make(map[string][]SequenceObject)
	pq := &PriorityQueue{}
	heap.Init(pq)

	NodeLock = sync.RWMutex{}
	PQLock = sync.RWMutex{}
	SequenceLock = sync.RWMutex{}

	removeFile("log.txt")
}

func main() {
	if len(os.Args) > 1 {
		configFilePath = os.Args[2]
	} else {
		log.Println("Please enter the node number and config file in the command line")
	}

	ReadFile(configFilePath)

	// listen on port
	listener, _ := net.Listen("tcp", ":" + hostNode.Port)
	// if err != nil {
	// 	log.Println("Failed to listen on port ", err)
	// }
	defer listener.Close()
	fmt.Println("Listen successfully")

	fmt.Println("Please wait for around 10 seconds for all the nodes to be connected")

	time.Sleep(10e9)

	initialize()

	for len(ConnectedNodes) < (nodeNum - 1) {
		for i := 1; i <= nodeNum; i++ {
			nodeId := "node" + strconv.Itoa(i)
			if nodeId == hostNode.Id {
				continue
			}
			nodeInfo := NodesToPorts[nodeId]
			
			conn, err := net.Dial("tcp", nodeInfo.Address + ":" + nodeInfo.Port)
			if err != nil {
				// fmt.Println("err ", err)
				continue
			}
	
			node := Node {
				Id: nodeId,
				Address: nodeInfo.Address,
				Port: nodeInfo.Port,
				Connection: conn,
			}
			ConnectedNodes[nodeId] = node
			fmt.Println("Successfully established connection with  ", nodeId)
			defer conn.Close()
		}
	}

	// time.Sleep(10e9)

	// send a new transaction
	go sendTransaction()


	for {
		// listen to other nodes
		conn, err := listener.Accept()

		if err != nil {
			// log.Println("Failed to receive message ", err)
			return
		}

		ip := conn.RemoteAddr().String()
		
		ipAddress := strings.Split(ip, ":")

		nodeId := AddressToId[ipAddress[0]]
		node := ConnectedNodes[nodeId]

		// go CheckConnection(&node)
		go receiveMsg(conn, node.Id)
	}
}
