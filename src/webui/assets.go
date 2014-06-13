package webui

import (
	"github.com/GeertJohan/go.rice"
	"../fs"
)

func read() string {
	box, err := rice.FindBox("assets")
	fs.CheckError(err)

	txt, err := box.String("data.txt")
	fs.CheckError(err)

	return txt
}
