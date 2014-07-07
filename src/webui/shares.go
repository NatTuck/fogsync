
package webui

import (
	"net/http"
	"encoding/json"
	"os"
	"io/ioutil"
	"../shares"
	"../config"
	"../fs"
)

type Shares struct {
	Shares []*shares.ShareConfig `json:"shares"`
}

func serveShares(ww http.ResponseWriter, req *http.Request) {
	elems := splitPath(req)

	if len(elems) == 1 {
		serveSharesIndex(ww, req)
	} else {
		serveShare(elems[1], ww, req)
	}

}

func serveSharesIndex(ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	cfgs := make([]*shares.ShareConfig, 0)
	for _, ss := range(shares.GetMgr().List()) {
		cfgs = append(cfgs, ss.Config)
	}

	resp := Shares{cfgs}

	data, err := json.MarshalIndent(&resp, "", "  ")
	fs.CheckError(err)

	ww.Write(data)
}

type FileData struct {
	Name string
	Type string
	Size uint64
}

type FileList struct {
	Files []*FileData `json:"files"`
}

func serveShare(name string, ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	ss := shares.GetMgr().Get(name)
	pp := req.URL.RawQuery

	if pp == "" {
		pp = "/"
	}

	temp := config.TempName()
	defer os.Remove(temp)

	info, err := ss.Trie.Get(pp, temp)
	fs.CheckError(err)

	if info.Type == eft.INFO_DIR {

	} else {
		
	}


	data, err := ioutil.ReadFile(temp)
	fs.CheckError(err)

	ww.Write(data)
}
