package main

import(
	"fmt"
	"container/heap"
	"os"
	"log"
	"bufio"
	"strconv"
	"strings"
	"net"
)

type Message struct{
	DeliverStatus bool  // true-delivered, false-not delivered
	Priority int 
	Sender string // which node sends the message
	Transaction string // content of transaction
	Type string // T(Transaction), PR(Priority Request), PC(Priority Confirmed)
}

type Node struct{
	Address string
	Port string
}

var Account map[string]int // user account

var NodesToPorts map[string]Node // the port number different nodes listen to (read from config file)

// read from command line
var node string
var configFilePath string

// heap interface -> priority queue
type PriorityQueue []Message
// methods of PriorityQueue
func (pq PriorityQueue) Len() int {
	return len(pq)
}
func (pq PriorityQueue) Less(i, j int) bool{
	return pq[i].Priority < pq[j].Priority
}
func (pq PriorityQueue) Swap(i, j int){
	pq[i], pq[j] = pq[j], pq[i]
}
func (pq *PriorityQueue) Pop() interface{}{
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}
func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(Message))
}
func (pq PriorityQueue) Update(msg Message){
	for i := 0; i < len(pq); i++{
		if pq[i].DeliverStatus == false && pq[i].Sender == msg.Sender && pq[i].Transaction == msg.Transaction {
			pq[i].Priority = msg.Priority
			pq[i].DeliverStatus = true
		}
	}
}

func ReadFile(path string){
	NodesToPorts = make(map[string]Node)

	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Read files failed")
	}

	buf := bufio.NewReader(f)
	line, err := buf.ReadString('\n')
	if err != nil{
		log.Fatal("Config file structure is uncorrect")
	}
	
	nodeNum, _ := strconv.Atoi(strings.TrimSpace(line))

	for d := 0; d < nodeNum; d++ {
		line, _ := buf.ReadString('\n')
		nodeInfo := strings.Split(line, " ")
		n := Node{strings.TrimSpace(nodeInfo[1]), strings.TrimSpace(nodeInfo[2])}
		NodesToPorts[nodeInfo[0]] = n
	}
}

func handleConnection(conn net.Conn){
	defer conn.Close()

	
}

func sendMsg(msg string, msgType string) {
	if msgType == "T"{
		for key, value := range NodesToPorts{
			if key != node{
				conn, err := net.Dial("tcp", value.Address + ":" + value.Port)
				if err != nil {
					log.Fatal("Connection Failed", err)
				}
				defer conn.Close()
				// get priority
				// TODO
				p := 0
				message := Message{false, p, node, msg, "T"}
				conn.Write([]byte(message))
			}
		}
	}else if msgType == "PR"{

	}else if msgType == "PC"{

	}
}


func main(){
	if len(os.Args) > 1{
		node = os.Args[1]
		configFilePath = os.Args[2]
	}else{
		log.Fatal("Please enter the node number and config file in the command line")
	}

	ReadFile(configFilePath)

	// Account = make(map[string]int) // initial map

	// m1 := Message{false, 2, 1, "1234", "T"}
	// m2 := Message{false, 5, 2, "124", "T"}
	// pq := &PriorityQueue{}
	// heap.Init(pq)

	// pq.Push(m1)
	// pq.Push(m2)

	// m3 := Message{false, 1, 2, "124", "CP"}
	// fmt.Println(*pq)

	// pq.Update(m3)

	// fmt.Println(*pq)
	

	// listen to port
	listener, err := net.Listen("tcp",":" + NodesToPorts[node].Port)
	if err != nil {
		log.Fatal("Error! ", err)
	}
	defer listener.Close()

	for{
		// listen
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal("Error", err)
		}

		go handleConnection(conn)

		// send
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			sendMsg(s.Text(), "T")
		}


	}


}