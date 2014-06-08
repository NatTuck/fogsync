
package cache

import (
	"encoding/binary"
	"encoding/hex"
	"../fs"
)

const BPTR_SIZE  = 44

type Bptr struct {
	Hash  []byte
	Byte0 uint32
	Byte1 uint32
	Depth uint32
}

func (bp *Bptr) Bytes() []byte {
	bytes := make([]byte, 44, 44)

	be := binary.BigEndian

	copy(bytes[0:32], bp.Hash)
	be.PutUint32(bytes[32:36], bp.Byte0)
	be.PutUint32(bytes[36:40], bp.Byte1)
	be.PutUint32(bytes[40:44], bp.Depth)

	return bytes
}

func (bp *Bptr) String() string {
	return hex.EncodeToString(bp.Bytes())
}

func BptrFromBytes(bytes []byte) Bptr {
	if len(bytes) != BPTR_SIZE {
		fs.PanicHere("Short bptr: " + hex.EncodeToString(bytes))
	}

	be := binary.BigEndian

	hash  := bytes[0:32]
	byte0 := be.Uint32(bytes[32:36])
	byte1 := be.Uint32(bytes[36:40])
	depth := be.Uint32(bytes[40:44])

	return Bptr{hash, byte0, byte1, depth}
}

func BptrFromString(text string) Bptr {
	bytes, err := hex.DecodeString(text)
	fs.CheckError(err)
	return BptrFromBytes(bytes)
}
