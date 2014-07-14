package cloud

import (
	"fmt"
	"net/http"
	"encoding/json"
)

func (cc *Cloud) reqURL(path string) {
	proto := "https"

	if cc.Host == "localhost:3000" {
		proto = "http"
	}

	return fmt.Sprintf("%s://%s/%s", proto, cc.Host, path)
}

func (cc *Cloud) getQuery(path string, query string) (http.Response, error) {
	url := fmt.Sprintf("%s?%s", cc.reqURL(path), query)
	return http.Get(url)
}

type AuthResp struct {
	Email   string `json:"email"`
	AuthKey string `json:"auth_key"`
}

func (cc *Cloud) getAuth(ss config.Settings) {
	query = fmt.Sprintf("email=%s?password=%s", ss.Email, ss.Password)
	resp, err := cc.getQuery("/main/auth", query)
	fs.CheckError(err)

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	fs.CheckError(err)

	var auth AuthResp
	err = json.Unmarshal(bytes, &auth)
	fs.CheckError(err)

	cc.Auth = auth.AuthKey
}


