package main

func main() {
	//port := 6379
	//args := os.Args
	//if len(args) >= 2 {
	//	port, _ = strconv.Atoi(args[2])
	//}
	//redisServer := initServer(port)
	//redisServer.run()
	cmdHandler := initCommandHandler()
	cmd := "*2\r\n$5\r\nHELLO\r\n*2\r\n$2\r\nHI\r\n$2\r\nHI\r\n"
	index := 0
	resp := cmdHandler.parseResp(cmd, &index)
	for i := 0; i < len(resp); i++ {
		println(resp[i])
	}
}
