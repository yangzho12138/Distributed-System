package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		var buf = make([]byte, 1024)

		n, err := conn.Read(buf) // n is the byte read from connection
		if err != nil {
			log.Fatal("Error! ", err)
		}

		message := string(buf[:n])
		timestamp := strings.Split(message, " ")
		current := strconv.FormatFloat(float64(time.Now().UnixNano())/float64(1000000000), 'f', 9, 64)

		// format of output: send time + receive time + bandwidth
		log.Println(timestamp[0] + " " + current + " " + strconv.Itoa(n)) // output -> log.txt

		fmt.Println(message)
	}
}

func deleteFile(path string){
	e := os.Remove(path)
    if e != nil {
        log.Fatal(e)
    }
}

func logFileInit(path string){
	_, err := os.Stat(path)
	if err != nil{
		if os.IsExist(err){
			deleteFile(path)
		}
	}
}

func logOutputInit(path string){
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}
	defer f.Close()
	log.SetOutput(f)
}

func main() {
	port := ":8080"
	if len(os.Args) > 1 {
		port = ":" +
			os.Args[1]
	} else {
		log.Fatal("Please enter the port number listening to in the command line")
	}

	// listen to port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error! ", err)
	}
	defer listener.Close()

	logFileInit("log.txt")
	logOutputInit("log.txt")
	
	fmt.Println("Listen to " + port + " port Success")

	// wait for connection
	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal("Error", err)
		}

		go handleConnection(conn)
	}

}
