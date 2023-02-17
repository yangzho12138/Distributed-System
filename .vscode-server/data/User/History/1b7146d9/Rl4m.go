package main

import (
	"net"
	"log"
	"os"
	"fmt"
	"time"
	"strings"
	"strconv"
)

func handleConnection(conn net.Conn){
	defer conn.Close()

	for{
		var buf = make([]byte, 1024)

		before := strconv.FormatFloat(float64(time.Now().UnixNano())/float64(1000000000), 'f', 9, 64)
		n, err := conn.Read(buf)
		after := strconv.FormatFloat(float64(time.Now().UnixNano())/float64(1000000000), 'f', 9, 64)

		message := string(buf[:n])
		if strings.Contains(message, "connected") {
			timestamp := strings.Split(message, " ")
			log.Println("Delay  " + timestamp[0] + " " + before + " " + after) // output -> log.txt
		}else{
			log.Println("Bandwidth " + strconv.Itoa(n) + " " + before + " " + after) output -> log.txt
		}

		if err != nil {
			log.Fatal("Error! ", err)
		}

		fmt.Println(message)
	}
}

func main(){
	port := ":8080"
	if len(os.Args) > 1 {
		port = ":" + 
		os.Args[1]
	}else {
		log.Fatal("Please enter the port number listening to in the command line")
	}

	// listen to port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error! ", err)
	}
	defer listener.Close()
	// log.Println("Listen to " + port + " port Success")

	// log output
	f, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil{
		return
	}
	defer f.Close()
	log.SetOutput(f)

	// wait for connection
	for{
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal("Error", err)
		}

		go handleConnection(conn)
	}

}