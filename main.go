package main

import (
	"os"
	"strconv"
)

func main() {
	port := 6379
	args := os.Args
	if len(args) >= 2 {
		port, _ = strconv.Atoi(args[2])
	}
	redisServer := initServer(port)
	redisServer.run()

}
