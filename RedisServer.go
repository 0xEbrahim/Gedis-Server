package main

import (
	"log"
	"net"
	"strconv"
)

type RedisServer struct {
	port     int
	running  bool
	listener net.Listener
	cmd      *CommandHandler
}

func initServer(port int) *RedisServer {
	return &RedisServer{
		port:     port,
		running:  false,
		listener: nil,
		cmd:      initCommandHandler(),
	}
}

func (rs *RedisServer) shutDown() {
	rs.running = false
	if rs.running {
		rs.listener = nil
	}
}

func (rs *RedisServer) run() {
	port := ":" + strconv.Itoa(rs.port)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error while listening to port ", port)
	}
	println("Gedis server is up and running on port ", port)
	rs.listener = ln
	rs.running = true
	for rs.running {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("Error with client connection")
		}
		go func(conn net.Conn) {
			defer func(conn net.Conn) {
				err := conn.Close()
				if err != nil {

				}
				println("Connection Closed")
			}(conn)
			b := make([]byte, 1024)
			for {
				n, err := conn.Read(b)
				if err != nil {
					break
				}
				response := rs.cmd.execCommand(string(b[:n]))
				_, err = conn.Write([]byte(response))
				if err != nil {
					log.Fatal("Error delivering the response")
				}
			}
		}(conn)
	}
}
