package webui

import (
	"net/http"
	"path"
	"github.com/GeertJohan/go.rice"
	"../fs"
)

func serveIndex(ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"text/html"}
	
	ww.Write(getAsset("index.html"))
}

func serveAssets(ww http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "" || req.URL.Path == "/" {
		serveIndex(ww, req)
		return
	}

	name := path.Base(req.URL.Path)

	ctype := "application/octet-stream"

	switch path.Ext(name) {
	case ".css":
		ctype = "text/css"
	case ".js":
		ctype = "application/javascript"
	default:
		// do nothing
	}
	
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{ctype}

	ww.Write(getAsset(name))
}

func getAsset(name string) []byte {
	box, err := rice.FindBox("assets/public")
	fs.CheckError(err)

	data, err := box.Bytes(name)
	fs.CheckError(err)

	return data
}
