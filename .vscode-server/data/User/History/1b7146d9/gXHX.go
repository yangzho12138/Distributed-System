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
	if len(os.Args) > 0 {
		port := os.Args[0]
	}
	// listen to port
	listener, err := net.Listen("tcp", ":"+port)
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