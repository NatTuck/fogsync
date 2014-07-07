package webui

import (
	"net/http"
	"strings"
)

func splitPath(req *http.Request) []string {
	path := req.URL.Path
	path = strings.Trim(path, "/")
	return strings.Split(path, "/")
}
