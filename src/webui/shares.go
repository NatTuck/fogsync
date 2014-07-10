
package webui

import (
	"net/http"
	"encoding/json"
	"strconv"
	"../eft"
	"../shares"
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
		id, err := strconv.Atoi(elems[1])
		fs.CheckError(err)

		serveShare(id, ww, req)
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

type FileInfo struct {
	Type string
	Size uint64
	ModT string
	Exec bool
	Hash string
	Path string
	MoBy string
}

type LongShare struct {
	Id    int    `json:"id"`
	Name  string
	Key   string
	Files []*FileInfo
}

func toFileInfo(info *eft.ItemInfo) *FileInfo {
	return &FileInfo{
		Type: info.TypeName(),
		Size: info.Size,
		ModT: info.DateText(),
		Exec: info.IsExec(),
		Hash: info.HashText(),
		Path: info.Path,
		MoBy: info.MoBy,
	}
}

type ShareWrapper struct {
	Share LongShare `json:"share"`
}

func serveShare(id int, ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	ss := shares.GetMgr().GetById(id)

	infos, err := ss.Trie.ListAllInfos()
	if err != nil {
		ww.WriteHeader(500)
		ww.Write([]byte("Error: " + err.Error()))
		fs.CheckError(err)
	}

	fis := make([]*FileInfo, 0)

	for _, info := range(infos) {
		fi := toFileInfo(info)
		fis = append(fis, fi)
	}

	files := ShareWrapper{
		Share: LongShare{
			Id:   ss.Config.Id,
			Key : ss.Config.Key,
			Name: ss.Config.Name,
			Files: fis,
		}, 
	}

	data, err := json.MarshalIndent(&files, "", "  ")
	fs.CheckError(err)

	ww.Write(data)
}
