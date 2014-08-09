package webui

import (
	"net/http"
	"../fs"
)

func checkError(ww http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	ww.WriteHeader(500)
	ww.Write([]byte("Error: " + err.Error()))
	fs.CheckError(err)
}
