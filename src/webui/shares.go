
package webui

import (
	"net/http"
	"encoding/json"
	"../shares"
	"../fs"
)

type Shares struct {
	Shares []*shares.Share `json:"shares"`
}

func serveShares(ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	shares := shares.GetMgr().List()
	resp := Shares{shares}

	data, err := json.MarshalIndent(&resp, "", "  ")
	fs.CheckError(err)

	ww.Write(data)
}

