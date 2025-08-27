package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

type RedisServer struct {
	port     int
	running  bool
	listener net.Listener
	cmd      *CommandHandler
	db       *Database
	wg       sync.WaitGroup
}

func initServer(port int) *RedisServer {
	return &RedisServer{
		port:     port,
		running:  false,
		listener: nil,
		cmd:      initCommandHandler(),
		db:       getDBInstance(),
	}
}

func (rs *RedisServer) shutDown() {
	if rs.running {
		rs.running = false
		if rs.listener != nil {
			rs.listener.Close()
		}
		rs.wg.Wait()
		rs.db.flush()
		rs.listener = nil
		rs.port = 0
	}
}

func (rs *RedisServer) run() {
	port := ":" + strconv.Itoa(rs.port)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error while listening to port ", port)
	}
	println("Gedis server is up and running on port ", port)
	go rs.CTX()
	rs.listener = ln
	rs.running = true
	for rs.running {
		conn, err := ln.Accept()
		if err != nil {
			if !rs.running {
				break
			}
			println("Accept error")
			continue
		}
		go rs.handler(conn)
	}
}

func (rs *RedisServer) handler(conn net.Conn) {
	rs.wg.Add(1)
	defer rs.wg.Done()
	defer conn.Close()
	b := make([]byte, 1024)
	for {
		n, err := conn.Read(b)
		if err != nil {
			break
		}
		response := rs.cmd.execCommand(string(b[:n]))
		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Println("Client write error:", err)
			return
		}
	}
}

func (rs *RedisServer) CTX() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	rs.shutDown()
	return

}
