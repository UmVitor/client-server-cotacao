package main

import (
	"client-server/client"
	"client-server/server"
	"time"
)

func main() {
	go server.InitServer()
	time.Sleep(1 * time.Second)
	client.InitClient()
}
