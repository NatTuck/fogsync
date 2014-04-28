package config

import (
	"strings"
	"path"
)

type SyncPath struct {
	rel_path string
}

func NewSyncPath(pp string) *SyncPath {
	sync := SyncDir()

	if strings.Index(pp, sync) == -1 {
		return &SyncPath{rel_path: pp}
	}

	start := len(sync)

	if pp[start:start + 1] == "/" {
		start += 1
	}

	sp := new(SyncPath)
	sp.rel_path = "" + pp[start:]

	return sp
}

func (pp *SyncPath) Full() string {
	sync := SyncDir()
	return path.Join(sync, pp.rel_path)
}

func (pp *SyncPath) Short() string {
	return pp.rel_path
}
