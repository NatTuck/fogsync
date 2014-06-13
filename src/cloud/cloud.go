package cloud

import (
	"../config"
)

type Cloud struct {
	Host string
	Auth string
}

func (cc *Cloud) Save() {
	clouds := make(map[string]Cloud, 0)
	err := config.GetObj("clouds", &clouds)
	clouds[cc.Host] = cc
	config.PutObj("clouds", &clouds)
}


