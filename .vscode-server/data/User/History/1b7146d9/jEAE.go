package main

import (
	"net"
	"log"
	"os"
)

func handleConnection(conn net.Conn){
	defer conn.Close()

	for{
		var buf = make([]byte, 1024)

		n, err := conn.Read(buf)

		if err != nil {
			log.Fatal("Error! ", err)
		}

		log.Println(string(buf[:n]))
	}
}

func main(){
	port := ":8080"
	if len(os.Args) > 0 {
		port = ":" + 
		os.Args[0]
	}else {
		log.Fatal("Please enter the port number listening to in the command line")
	}
	log.Println(port)
	// listen to port
	// listener, err := net.Listen("tcp", port)
	// if err != nil {
	// 	log.Fatal("Error! ", err)
	// }
	// defer listener.Close()
	// log.Println("Listen to" + port + "port Success")

	// // wait for connection
	// for{
	// 	conn, err := listener.Accept()

	// 	if err != nil {
	// 		log.Fatal("Error", err)
	// 	}

	// 	go handleConnection(conn)
	// }

}