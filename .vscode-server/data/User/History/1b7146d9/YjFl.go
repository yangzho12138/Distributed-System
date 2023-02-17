package main

import (
	"net"
	"log"
)

func main(){
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error", err)
	}
	defer listener.Close()
	log.Println("Success")
}