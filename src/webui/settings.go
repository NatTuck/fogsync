package webui

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"../shares"
	"../config"
	"../fs"
)

func serveSettings(ww http.ResponseWriter, req *http.Request) {
	if req.Method == "PUT" {
		saveSettings(ww, req)
		return
	}

	hdrs := ww.Header()
	hdrs["Content-Type"] = []string{"application/json"}

	settings := config.GetSettings()

	data, err := json.MarshalIndent(&settings, "", "  ")
	fs.CheckError(err)

	ww.Write(data)
}

func saveSettings(ww http.ResponseWriter, req *http.Request) {
	bytes, err := ioutil.ReadAll(req.Body)
	fs.CheckError(err)

	settings := config.Settings{}
	err = json.Unmarshal(bytes, &settings)
	fs.CheckError(err)

	settings.Save()

	ww.WriteHeader(204)

	fmt.Println("Saved settings")

	shares.Reload()
}
