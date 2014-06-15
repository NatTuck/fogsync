
package webui

import (
	"net/http"
	"encoding/json"
	"../config"
	"../fs"
)

type Shares struct {
	Shares []config.Share `json:"shares"`
}

func serveShares(ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	shares := config.Shares()
	resp := Shares{shares}

	data, err := json.MarshalIndent(&resp, "", "  ")
	fs.CheckError(err)

	ww.Write(data)
}

