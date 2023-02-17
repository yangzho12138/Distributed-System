package main

import (
	"net"
	"log"
)

func handleConnection(conn net.Conn){
	defer conn.Close()

	for{
		var buf = make([]byte, 1024)

		n, err := conn.Read(buf)

		if err != nil {
			log.Fatal(err)
		}

		log.Println(string(buf[:n]))
	}
}

func main(){
	// listen to 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error", err)
	}
	defer listener.Close()
	log.Println("Listen to 8080 port Success")

	// wait for connection
	for{
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal("Error", err)
		}

		go handleConnection(conn)
	}

}