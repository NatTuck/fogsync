
package cache

import (
	"encoding/json"
	"../fs"
)

type Dir struct {
	Name string
	Ents map[string]DirEnt
}

type DirEnt struct {
	Type string
	Bptr string
	Size int64
	Hash string
	Exec bool   // Regular files only
	Link string // Symlinks only
}

func DirFromJson(text []byte) Dir {
	dd  := Dir{}
	err := json.Unmarshal(text, &dd)
	fs.CheckError(err)
	return dd
}

func (dd *Dir) Json() []byte {
	json, err := json.MarshalIndent(dd, "", "  ")
	fs.CheckError(err)
	return json
}

