package webui

import (
	"net/http"
	"encoding/json"
	"../config"
	"../fs"
)

func Start() {
	go serverLoop()
}

func serverLoop() {
	http.HandleFunc("/", serveAssets)
	http.HandleFunc("/shares/", serveShares)
	http.HandleFunc("/settings/", serveSettings)
	http.HandleFunc("/about/", serveAbout)

	http.ListenAndServe(":5000", nil)
}

type AboutInfo struct {
	Version string
}

func serveAbout(ww http.ResponseWriter, req *http.Request) {
	ainfo := &AboutInfo{
		Version: config.VERSION,
	}

	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	data, err := json.MarshalIndent(&ainfo, "", "  ")
	fs.CheckError(err)

	ww.Write(data)
}
