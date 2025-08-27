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
}

func initServer(port int) *RedisServer {
	return &RedisServer{
		port: port,
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
	println("Gedis server is up and running")
	rs.listener = ln
	rs.running = true
}
