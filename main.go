package main

import (
	"log"
	"time"

	"lcc-go/server"
	"lcc-go/utils"
)

func main() {
	server.StartServer()

	time.Sleep(800 * time.Millisecond)

	ip := utils.GetLocalIP()
	url := "http://" + ip + ":3000/#/host"
	if err := utils.OpenBrowser(url); err != nil {
		log.Println("Gagal membuka browser:", err)
	}

	select {}
}
