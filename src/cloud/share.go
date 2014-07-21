package cloud

import (
	"encoding/json"
	"path"
	"fmt"
	"time"
	"../fs"
)

type ShareInfo struct {
	Name       string `json:"name"`
	Root       string `json:"root"`
	BlockSize  int64  `json:"block_size"`
	BlockCount int64  `json:"block_count"`
	TransBytes int64  `json:"trans_bytes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (cc *Cloud) GetShare(name_hmac string) (*ShareInfo, error) {
	cpath := path.Join("/shares", name_hmac)
	data, err := cc.getJSON(cpath)
	if err != nil {
		return nil, fs.Trace(err)
	}

	fmt.Println("XX - Response", string(data))

	sinfo := &ShareInfo{}
	err = json.Unmarshal(data, sinfo)
	if err != nil {
		return nil, fs.Trace(err)
	}

	return sinfo, nil
}
