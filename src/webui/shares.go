
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

type SharesList struct {
	Shares []*shares.ShareConfig
	Broken []string
}

func serveSharesIndex(ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	mgr  := shares.GetMgr()
	list := &SharesList{
		Shares: mgr.ListConfigs(),
		Broken: mgr.ListBroken(),
	}

	data, err := json.MarshalIndent(&list, "", "  ")
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

	infos, err := ss.Trie.ListInfos()
	if err != nil {
		ww.WriteHeader(500)
		ww.Write([]byte("Error: " + err.Error()))
		fs.CheckError(err)
	}

	fis := make([]*FileInfo, 0)

	for _, info := range(infos) {
		fi := toFileInfo(&info)
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
