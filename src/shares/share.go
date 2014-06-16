package shares

import (
	"sync"
)

type Share struct {
	Mutex sync.Mutex
}
