package cloud

import (
	"fmt"
	"path"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"../config"
	"../fs"
)

type ErrorJSON struct {
	Error string `json:'error'`
}

func (cc *Cloud) reqURL(cpath string) string {
	proto := "https"

	if cc.Host == "localhost:3000" {
		proto = "http"
	}

	return fmt.Sprintf("%s://%s%s", proto, cc.Host, path.Join("/", cpath))
}

func (cc *Cloud) getJSON(cpath string) ([]byte, error) {
	req, err := http.NewRequest("GET", cc.reqURL(cpath), nil)
	if err != nil {
		return nil, fs.Trace(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-FogSync-Auth", cc.Auth)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fs.Trace(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fs.Trace(err)
	}

	if resp.StatusCode != 200 {
		msg := &ErrorJSON{}
		err := json.Unmarshal(data, msg)
		if err != nil {
			return nil, fmt.Errorf("GET failed: %s", resp.Status)
		} else {
			return nil, fmt.Errorf("GET failed: %s", msg.Error)
		}
	}

	return data, nil
}

func (cc *Cloud) postJSON(cpath string, data []byte) ([]byte, error) {
	panic("TODO")
}

func (cc *Cloud) postFile(cpath string, file_path string) error {
	panic("TODO")
}

func (cc *Cloud) getQuery(path string, query string) (*http.Response, error) {
	url := fmt.Sprintf("%s?%s", cc.reqURL(path), query)
	return http.Get(url)
}

type AuthResp struct {
	Email   string `json:"email"`
	AuthKey string `json:"auth_key"`
}

func (cc *Cloud) getAuth(ss config.Settings) error {
	query := fmt.Sprintf("email=%s&password=%s", ss.Email, ss.Passwd)
	resp, err := cc.getQuery("/main/auth", query)
	if err != nil {
		return fs.Trace(err)
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fs.Trace(err)
	}

	var auth AuthResp
	err = json.Unmarshal(bytes, &auth)
	if err != nil {
		return fs.Trace(err)
	}

	cc.Auth = auth.AuthKey

	return nil
}


