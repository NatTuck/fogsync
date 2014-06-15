package webui

import (
	"net/http"
)

func Start() {
	http.HandleFunc("/", serveAssets)
	http.HandleFunc("/shares/", serveShares)
	http.HandleFunc("/settings/", serveSettings)

	http.ListenAndServe(":5000", nil)
}



