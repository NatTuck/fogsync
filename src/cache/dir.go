
package cache

import (
	"encoding/json"
	"../fs"
	"../pio"
)

type Dir map[string]DirEnt

type DirEnt struct {
	Type string
	Bptr string
	Size int64
	Hash string
	Host string // Last modified where
	Mtime int64 // Last modified when 
	Exec bool   // Regular files only
	Link string // Symlinks only
}

func EmptyDir() Dir {
	return Dir(make(map[string]DirEnt))
}

func DirFromJson(text []byte) Dir {
	dd  := Dir{}
	err := json.Unmarshal(text, &dd)
	fs.CheckError(err)
	return dd
}

func DirFromFile(name string) Dir {
	text := pio.ReadFile(name)
	return DirFromJson(text)
}

func (dd *Dir) Json() []byte {
	json, err := json.MarshalIndent(dd, "", "  ")
	fs.CheckError(err)
	return json
}

func (ent *DirEnt) GetBptr() Bptr {
	return BptrFromString(ent.Bptr)
}
