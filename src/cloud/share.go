package cloud

import (
	"encoding/json"
	"path"
	"fmt"
	"time"
	"../fs"
	"../eft"
)

var FULL_BLOCK_SIZE = int64(eft.BLOCK_SIZE + eft.BLOCK_OVERHEAD)

type ShareInfo struct {
	NameHmac   string `json:"name"`
	Root       string `json:"root"`
	Secrets    string `json:"secrets"`
	BlockSize  int64  `json:"block_size"`
	BlockCount int64  `json:"block_count"`
	TransBytes int64  `json:"trans_bytes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (cc *Cloud) GetShares() ([]ShareInfo, error) {
	data, err := cc.getJSON("/shares")
	if err != nil {
		return nil, fs.Trace(err)
	}

	sinfos := make([]ShareInfo, 0)

	err = json.Unmarshal(data, &sinfos)
	if err != nil {
		return nil, fs.Trace(err)
	}

	return sinfos, nil
}

func (cc *Cloud) GetShare(name_hmac string) (*ShareInfo, error) {
	cpath := path.Join("/shares", name_hmac)
	data, err := cc.getJSON(cpath)
	if err != nil {
		return nil, err // Maybe ErrNotFound
	}

	sinfo := &ShareInfo{}
	err = json.Unmarshal(data, sinfo)
	if err != nil {
		return nil, fs.Trace(err)
	}

	if sinfo.BlockSize != FULL_BLOCK_SIZE {
		return nil, fmt.Errorf("Bad remote block size: %d", sinfo.BlockSize)
	}

	return sinfo, nil
}

type ShareCreate struct {
	NameHmac  string `json:"name"`
	Secrets   string `json:"secrets"`
	BlockSize int64  `json:"block_size"`
}

func (cc *Cloud) CreateShare(name_hmac string, secrets string) (*ShareInfo, error) {
	req_obj := &ShareCreate{
		NameHmac:  name_hmac,
		Secrets:   secrets,
		BlockSize: int64(eft.BLOCK_SIZE + eft.BLOCK_OVERHEAD),
	}
	req_data, err := json.Marshal(req_obj)
	if err != nil {
		return nil, fs.Trace(err)
	}

	resp, err := cc.postJSON("/shares", req_data)
	if err != nil {
		return nil, fs.Trace(err)
	}

	sinfo := &ShareInfo{}
	err = json.Unmarshal(resp, sinfo)
	if err != nil {
		return nil, fs.Trace(err)
	}

	return sinfo, nil
}

func (cc *Cloud) DeleteShare(name_hmac string) error {
	cpath := fmt.Sprintf("/shares/%s", name_hmac)
	resp, err := cc.httpRequest("DELETE", cpath, nil)
	if err != nil {
		return fs.Trace(err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("HTTP %s", resp.Status)
	}

	return nil
}

func (cc *Cloud) FetchBlocks(name_hmac string, src_path string, dst_path string) error {
	cpath := fmt.Sprintf("/shares/%s/get", name_hmac)
	err := cc.postFile(cpath, src_path, dst_path)
	if err != nil {
		return fs.Trace(err)
	}

	return nil
}

func (cc *Cloud) SendBlocks(name_hmac string, src_path string) error {
	cpath := fmt.Sprintf("/shares/%s/put", name_hmac)
	err := cc.postFile(cpath, src_path, "")
	if err != nil {
		return fs.Trace(err)
	}

	return nil
}

func (cc *Cloud) RemoveList(name_hmac string, src_path string) error {
	cpath := fmt.Sprintf("/shares/%s/remove", name_hmac)
	err := cc.postFile(cpath, src_path, "")
	if err != nil {
		return fs.Trace(err)
	}

	return nil
}

type ShareSwapRoot struct {
	Prev string `json:"prev"`
	Root string `json:"root"`
}

func (cc *Cloud) SwapRoot(name_hmac string, prev string, root string) error {
	fmt.Println("XX - Swapping root", prev, " -> ", root)

	req_obj := &ShareSwapRoot{
		Prev: prev,
		Root: root,
	}
	req_data, err := json.Marshal(req_obj)
	if err != nil {
		return fs.Trace(err)
	}

	cpath := fmt.Sprintf("/shares/%s/casr", name_hmac)
	_, err = cc.postJSON(cpath, req_data)
	if err != nil {
		return fs.Trace(err)
	}

	return nil
}

