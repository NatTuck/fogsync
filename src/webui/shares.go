
package webui

import (
	"net/http"
	"encoding/json"
	"fmt"
	"../eft"
	"../shares"
	"../cloud"
	"../fs"
)

func serveShares(ww http.ResponseWriter, req *http.Request) {
	elems := splitPath(req)

	fmt.Println("XX - serveShares", req.Method, elems)

	if len(elems) == 1 {
		switch req.Method {
		case "GET":
			getSharesIndex(ww, req)
		case "POST":
			createShare(ww, req)
		default:
			fs.PanicHere("Bad method: " + req.Method)
		}
	} else {
		switch req.Method {
		case "GET":
			getShare(elems[1], ww, req)
		case "DELETE":
			delShare(elems[1], ww, req)
		default:
			fs.PanicHere("Bad method: " + req.Method)
		}
	}
}

type SharesList struct {
	Shares []*shares.ShareConfig
	Broken []string
}

func getSharesIndex(ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	list := &SharesList{
		Shares: shares.ListConfigs(),
		Broken: shares.ListBroken(),
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

func getShare(name string, ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	ss := shares.Get(name)

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

func delShare(name string, ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	cc, err := cloud.New()
	checkError(ww, err)

	err = cc.DeleteShare(name)
	checkError(ww, err)
	
	ww.WriteHeader(204)
}

func createShare(ww http.ResponseWriter, req *http.Request) {
	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	err := req.ParseForm()
	checkError(ww, err)

	share_name := req.Form.Get("name")

	fmt.Println("XX - Create Share", share_name)

	shares.Create(share_name)

	http.Redirect(ww, req, "/#/shares", 303)
}
