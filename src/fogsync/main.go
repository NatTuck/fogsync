package main

import (
	"fmt"
	"time"
	"../fs"
	"../webui"
	"../shares"
	"../config"
)

func main() {
	fmt.Println("Starting Web UI...")
	webui.Start()
	
	time.Sleep(1 * time.Second)

	settings := config.GetSettings()
	if !settings.Ready() {
		url := "http://localhost:5000/#/settings"
		err := fs.Launch(url)
		fs.CheckError(err)
	}

	shares.Reload()

	url := "http://localhost:5000/"
	fmt.Println("FogSync Started, visit", url)

	select {}
}

