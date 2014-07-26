package cloud

import (
	"fmt"
	"path"
	"bytes"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"io"
	"os"
	"../config"
	"../fs"
)

var ErrNotFound = fmt.Errorf("HTTP 404 Not Found (api)")

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

func (cc *Cloud) httpRequest(mm string, cpath string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(mm, cc.reqURL(cpath), body)
	if err != nil {
		return nil, fs.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-FogSync-Auth", cc.Auth)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fs.Trace(err)
	}

	return resp, nil
}

func (cc *Cloud) httpReqObj(mm string, cpath string, body io.Reader, obj interface{}) error {
	resp, err := cc.httpRequest(mm, cpath, body)

	if err != nil {
		return fs.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return ErrNotFound
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fs.Trace(err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := &ErrorJSON{}
		err := json.Unmarshal(data, msg)
		if err != nil {
			return fmt.Errorf("HTTP %s (non-json)", resp.Status)
		} else {
			return fmt.Errorf("HTTP %s: %s", resp.Status, msg.Error)
		}
	}

	err = json.Unmarshal(data, obj)
	if err != nil {
		return fs.Trace(err)
	}

	return nil
}


func (cc *Cloud) getJSON(cpath string) ([]byte, error) {
	resp, err := cc.httpRequest("GET", cpath, nil)
	if err != nil {
		return nil, fs.Trace(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fs.Trace(err)
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, ErrNotFound
		}

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

func (cc *Cloud) postJSON(cpath string, post_data []byte) ([]byte, error) {
	resp, err := cc.httpRequest("POST", cpath, bytes.NewBuffer(post_data))
	if err != nil {
		return nil, fs.Trace(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fs.Trace(err)
	}

	if resp.StatusCode == 404 {
		return data, ErrNotFound
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return data, fmt.Errorf("HTTP %s", resp.Status)
	}

	return data, nil
}

func (cc *Cloud) postFile(cpath string, file_path string) error {
	body, err := os.Open(file_path)
	if err != nil {
		return fs.Trace(err)
	}
	defer body.Close()

	req, err := http.NewRequest("POST", cc.reqURL(cpath), body)
	if err != nil {
		return fs.Trace(err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-FogSync-Auth", cc.Auth)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return fs.Trace(err)
	}

	if resp.StatusCode == 404 {
		return ErrNotFound
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("HTTP %s", resp.Status)
	}
	
	return nil
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


