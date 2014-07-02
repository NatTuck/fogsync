
package webui

import (
	"net/http"
	"encoding/json"
	"fmt"
	"../shares"
	"../fs"
)

type Shares struct {
	Shares []*shares.ShareConfig `json:"shares"`
}

func serveShares(ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	fmt.Println("XX - Serving shares")

	cfgs := make([]*shares.ShareConfig, 0)
	for _, ss := range(shares.GetMgr().List()) {
		fmt.Println("XX - Appending share", ss)
		cfgs = append(cfgs, ss.Config)
	}

	resp := Shares{cfgs}

	data, err := json.MarshalIndent(&resp, "", "  ")
	fs.CheckError(err)

	ww.Write(data)
}

