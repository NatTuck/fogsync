package cloud

import (
	"../config"
)

type Cloud struct {
	Host string
	Auth string
}

func Connect() *Cloud {
	cloud := &Cloud{}
	cloud.Load()

	// If we have auth data, try it.
	// If not or if it doesn't work, relog with email / password.
}

func (cc *Cloud) Load() {

}

func (cc *Cloud) Save() {
	clouds := make(map[string]Cloud, 0)
	err := config.GetObj("clouds", &clouds)
	clouds[cc.Host] = cc
	config.PutObj("clouds", &clouds)
}

