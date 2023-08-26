package main

import (
	"fmt"
	"log"
	"net"
)

const UpstreamPort = 5432
const ProxyPort = 5433

func main() {
	addr := fmt.Sprintf(":%d", ProxyPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening for connections on %d", ProxyPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
		}

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()

	log.Printf("Accepted connection")

	// TODO: read from connection until it ends
	// TODO: write read data over upstream connection. can test with netcat for now
}
