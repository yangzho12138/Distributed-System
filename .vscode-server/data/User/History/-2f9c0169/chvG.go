package main

import(
	"fmt"
	"container/heap"
	"os"
	"log"
	"bufio"
	"strconv"
	"strings"
)

type Message struct{
	DeliverStatus bool  // true-delivered, false-not delivered
	priority int 
	sender int // which node sends the message
}

type Node struct{
	IpAddress string
	Port string
}

var Account map[string]int // user account

var NodesToPorts map[string]Node // the port number different nodes listen to (read from config file)

// heap -> priority queue
type PriorityQueue []Message
// methods of PriorityQueue
func (pq PriorityQueue) Len() int {
	return len(pq)
}
func (pq PriorityQueue) Less(i, j int) bool{
	return pq[i].priority < pq[j].priority
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

func ReadFile(path string){
	NodesToPorts = make(map[string]Node)

	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Read files failed")
	}

	buf := bufio.NewReader(f)
	line, err := buf.ReadString('\n')
	fmt.Println(line)
	if err != nil{
		log.Fatal("Config file structure is uncorrect")
	}
	
	nodeNum, _ := strconv.Atoi(line)
	fmt.Println(nodeNum)

	for d := 0; d < nodeNum; d++ {
		line, _ := buf.ReadString('\n')
		fmt.Println(line)
		nodeInfo := strings.Split(line, " ")
		n := Node{nodeInfo[1], nodeInfo[2]}
		NodesToPorts[nodeInfo[0]] = n
	}
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
	fmt.Println(node)
	fmt.Println(NodesToPorts["node1"])
	fmt.Println(NodesToPorts["node2"])
	fmt.Println(NodesToPorts["node3"])


	Account = make(map[string]int) // initial map

	pq := &PriorityQueue{}
	heap.Init(pq)
	fmt.Println(*pq)

}