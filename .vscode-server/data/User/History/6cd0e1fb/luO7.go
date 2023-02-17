package main

import {
	"fmt"
	"log"
	"net"
}

func main(){
	fmt.Println("Logger")
	conn, err = net.Dail("tcp", "www.baidu.com:80")
	if err != nil{
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("success")
}