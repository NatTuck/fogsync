package shares

import (
	"fmt"
	"time"
	"../fs"
	"../pio"
	"../config"
)

var upload_delay = 5 * time.Second

func (ss *Share) upload() {
	ss.Uploads <-true
}

func (ss *Share) uploadLoop() {
	delay := time.NewTimer(upload_delay)

	for {
		select {
		case again := <-ss.Uploads:
			if again {
				delay.Reset(upload_delay)
			} else {
				fmt.Println("Shutting down uploadLoop")
				break
			}
		case _ = <-delay.C:
			ss.reallyUpload()
		}
	}
}

func (ss *Share) reallyUpload() {
	settings := config.GetSettings()
	if !settings.Ready() {
		fmt.Println("Skipping upload, no cloud configured.")
		return
	}

	// Check Cloud Share Setup
	cloud.GetShare(ss.HashedName())

	// Download

	cp, err := ss.Trie.MakeCheckpoint()
	fs.CheckError(err)
	defer cp.Cleanup()

	/*
	fmt.Println("== Added Blocks ==")
	addsData := pio.ReadFile(cp.Adds)
	fmt.Println(string(addsData))

	fmt.Println("== Dead Blocks ==")
	deadData := pio.ReadFile(cp.Dels)
	fmt.Println(string(deadData))
	*/
}
