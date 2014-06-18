package eft

const (
	INFO_FILE = 2
	INFO_DIR  = 3
	INFO_LINK = 4
)

type ItemInfo struct {
	Type int32
	Size uint64
	ModT uint64
	Mode uint32
	Hash []byte
	Path string
	MoBy string // last modified by (user@host)
}


