package main

import (
	"fmt"
	"time"
	"../shares"
	"../webui"
)

func main() {
	fmt.Println("Starting Shares...")
	mgr := shares.GetMgr()
	fmt.Println("XX - Starting scanner")
	mgr.ScanAll()

	fmt.Println("Starting Web UI...")
	webui.Start()

	time.Sleep(1 * time.Second)

	URL := "http://localhost:5000"
	fmt.Println("Visit", URL)

	fmt.Println("Startup complete")

	select {}
}

