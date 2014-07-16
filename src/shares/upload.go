package shares

import (
	"fmt"
	"time"
	"os"
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

	adds, dead, err := ss.Trie.Changes()
	fs.CheckError(err)
	defer os.Remove(adds)
	defer os.Remove(dead)

	fmt.Println("== Added Blocks ==")
	addsData := pio.ReadFile(adds)
	fmt.Println(string(addsData))

	fmt.Println("== Dead Blocks ==")
	deadData := pio.ReadFile(dead)
	fmt.Println(string(deadData))
}
