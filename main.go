package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {

	// Setup server for all connections.
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer ln.Close()

	// We need channels for holding all connections, incoming connections, dead connections and messages.
	var (
		aconns = make(map[net.Conn]int)
		conns  = make(chan net.Conn)
		dconns = make(chan net.Conn)
		msgs   = make(chan string)
		i      int
	)

	// Goroutine for accept incoming connections.
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatalln(err.Error())
			}
			conns <- conn
		}
	}()

	for {
		select {
		// Accept incoming connections.
		case conn := <-conns:
			aconns[conn] = i
			i++

			// Once we have the connection, we start reading message from it.
			go func(conn net.Conn, i int) {
				rd := bufio.NewReader(conn)
				for {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					msgs <- fmt.Sprintf("Client %v: %v", i, m)
				}

				// This client close their connection.
				dconns <- conn
			}(conn, i)
		case msg := <-msgs:

			// We have to broadcast it to all connections.
			for conn := range aconns {
				conn.Write([]byte(msg))
			}
		case dconn := <-dconns:
			log.Printf("Client %v was gone\n", aconns[dconn])
			delete(aconns, dconn)
		}
	}
}
