package eft

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"path"
	"sort"
	"fmt"
	"os"
)

// A directory maps names to ItemInfo types.
type Directory map[string]uint32

func (eft *EFT) getDir(dpath string) (ItemInfo, Directory, error) {
	temp := eft.TempName()
	defer os.Remove(temp)

	info, err := eft.getItem(dpath, temp)
	if err == ErrNotFound {
		info = ItemInfo{
			Path: dpath,
			Type: INFO_DIR,
			Size: 0,
		}

		return info, make(map[string]uint32), nil
	}
	if err != nil {
		return info, nil, trace(err)
	}

	if info.Type != INFO_DIR {
		return info, nil, trace(fmt.Errorf("getDir on non-directory item"))
	}

	text, err := ioutil.ReadFile(temp)
	if err != nil {
		return info, nil, trace(err)
	}
	
	if !BytesEqual(info.Hash[:], HashSlice(text)) {
		return info, nil, fmt.Errorf("Directory hash didn't match")
	}
	
	dir := make(map[string]uint32)
	
	err = json.Unmarshal(text, &dir)
	if err != nil {
		fmt.Println(string(text))
		return info, nil, trace(err)
	}

	return info, dir, nil
}

func (eft *EFT) putDir(info ItemInfo, dir Directory) error {
	temp := eft.TempName()
	defer os.Remove(temp)

	text, err := json.MarshalIndent(&dir, "", "  ")
	if err != nil {
		return trace(err)
	}

	dhash := HashSlice(text)
	copy(info.Hash[:], dhash)
	info.Size = uint64(len(text))

	err = ioutil.WriteFile(temp, text, 0600)
	if err != nil {
		return trace(err)
	}

	err = eft.putItem(info, temp)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) putParent(info ItemInfo) error {
	if info.Path == "/" {
		return nil
	}

	dpath, name := path.Split(info.Path)
	dpath = strings.TrimRight(dpath, "/")

	if dpath == "" {
		dpath = "/"
	}

	dinfo, dir, err := eft.getDir(dpath)
	if err != nil {
		return trace(err)
	}

	dir[name] = info.Type

	err = eft.putDir(dinfo, dir)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) ListDir(dpath string) ([]string, error) {
	eft.begin()

	_, dir, err := eft.getDir(dpath)
	if err != nil {
		eft.abort()
		return nil, trace(err)
	}

	names := make([]string, 0)

	for kk, _ := range(dir) {
		names = append(names, kk)
	}

	sort.Strings(names)

	items := make([]string, 0)

	for _, name := range(names) {
		ipath := path.Join(dpath, name)
		
		info, _, err := eft.getTree(ipath)
		if err != nil {
			eft.abort()
			return nil, trace(err)
		}

		desc := fmt.Sprintf("%s (%s) %d", name, info.TypeName(), info.Size)
		items = append(items, desc)
	}

	eft.commit()

	return items, nil
}

