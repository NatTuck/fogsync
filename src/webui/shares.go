
package webui

import (
	"net/http"
	"encoding/json"
	"../eft"
	"../shares"
	"../fs"
)

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

	data, err := json.MarshalIndent(&cfgs, "", "  ")
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

func serveShare(name string, ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	ss := shares.GetMgr().Get(name)

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

	share := LongShare{
		Key : ss.Config.Key,
		Name: ss.Config.Name,
		Files: fis,
	} 

	data, err := json.MarshalIndent(&share, "", "  ")
	fs.CheckError(err)

	ww.Write(data)
}
