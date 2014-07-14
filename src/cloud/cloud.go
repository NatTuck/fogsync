package cloud

import (
	"fmt"
	"../fs"
	"../config"
)

type Cloud struct {
	Host string
	Auth string
}

func New() *Cloud {
	ss := config.GetSettings()

	cc := &Cloud{
		Host: ss.Cloud,
	}

	cc.load()

	if cc.Auth == "" {
		cc.getAuth(ss)
		cc.save()
	}

	return cc
}

func (cc *Cloud) load() {
	cfg := fmt.Sprintf("clouds/%s", cc.Host)
	err := config.GetObj(cfg, cc)
	if err != nil {
		cc.Auth = ""
		fmt.Printf("No auth token for host %s\n", cc.Host)
	}
}

func (cc *Cloud) save() {
	cfg := fmt.Sprintf("clouds/%s", cc.Host)
	err := config.PutObj(cfg, cc)
	fs.CheckError(err)
}


