package webui

import (
	"net/http"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"../config"
	"../fs"
)

type Settings struct {
	Settings []config.Settings `json:"settings"`
}

type Setting struct {
	Setting config.Settings `json:"setting"`
}

func serveSettings(ww http.ResponseWriter, req *http.Request) {
	if req.Method == "PUT" {
		saveSettings(ww, req)
		return
	}

	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	settings := config.GetSettings()
	resp := Settings{[]config.Settings{settings}}

	data, err := json.MarshalIndent(&resp, "", "  ")
	fs.CheckError(err)

	ww.Write(data)
}

func saveSettings(ww http.ResponseWriter, req *http.Request) {
	bytes, err := ioutil.ReadAll(req.Body)
	fs.CheckError(err)

	settings := Setting{}
	err = json.Unmarshal(bytes, &settings)
	fs.CheckError(err)

	settings.Setting.Save()

	fmt.Println(settings)

	ww.WriteHeader(204)

	return
}
