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
	Sender int // which node sends the message
	Transaction string // content of transaction
	Type string // T(Transaction), PR(Priority Request), PC(Priority Confirmed)
}

type Node struct{
	Address string
	Port string
}

var Account map[string]int // user account

var NodesToPorts map[string]Node // the port number different nodes listen to (read from config file)

// heap interface -> priority queue
type PriorityQueue []*Message
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
func (pq *PriorityQueue) Update(msg Message){
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
		n := Node{nodeInfo[1], nodeInfo[2]}
		NodesToPorts[nodeInfo[0]] = n
	}
}

func handleConnection(conn net.Coon){
	defer conn.Close()

	
}

func sendMsg(msg string, msgType string) {

}


func main(){
	var node string
	var configFilePath string
	if len(os.Args) > 1{
		node = os.Args[1]
		configFilePath = os.Args[2]
	}else{
		log.Fatal("Please enter the node number and config file in the command line")
	}

	ReadFile(configFilePath)
	

	// listen to port
	listener, err := net.Listen("tcp", NodesToPorts[node].Port)
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

	}

	Account = make(map[string]int) // initial map

	pq := &PriorityQueue{}
	heap.Init(pq)
	fmt.Println(*pq)

}