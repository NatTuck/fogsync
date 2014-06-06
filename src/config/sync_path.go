package config

import (
	"strings"
	"path"
)

type SyncPath struct {
	Share    *Share
	rel_path string
}

func (ss *Share) NewSyncPath(pp string) SyncPath {
	sync := ss.Path()

	if strings.Index(pp, sync) == -1 {
		return SyncPath{rel_path: pp, Share: ss}
	}

	start := len(sync)

	if pp[start:start + 1] == "/" {
		start += 1
	}

	sp := SyncPath{}
	sp.rel_path = "" + pp[start:]
	sp.Share = ss

	return sp
}

func (pp *SyncPath) Full() string {
	return path.Join(pp.Share.Path(), pp.rel_path)
}

func (pp *SyncPath) Short() string {
	return pp.rel_path
}

