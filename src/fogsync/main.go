package main

import (
	"../webui"
	"../config"
)

func main() {
	// Handle first startup
	shares

	webui.Start()
}

