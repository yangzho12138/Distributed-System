package main

import (
	"log"
	"net"
)

func main(){
	conn, err := net.Dial("tcp", "www.baidu.com:80")
	if err != nil{
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("success")
}