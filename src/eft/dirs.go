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

func (eft *EFT) getDir(snap *Snapshot, dpath string) (ItemInfo, Directory, error) {
	temp := eft.TempName()
	defer os.Remove(temp)

	info, err := eft.getItem(snap, dpath, temp)
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
	
	if !HashesEqual(info.Hash, HashSlice(text)) {
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

func (eft *EFT) putDir(snap *Snapshot, info ItemInfo, dir Directory) error {
	temp := eft.TempName()
	defer os.Remove(temp)

	text, err := json.MarshalIndent(&dir, "", "  ")
	if err != nil {
		return trace(err)
	}

	info.Hash = HashSlice(text)
	info.Size = uint64(len(text))

	err = ioutil.WriteFile(temp, text, 0600)
	if err != nil {
		return trace(err)
	}

	err = eft.putItem(snap, info, temp)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) putParent(snap *Snapshot, info ItemInfo) error {
	if info.Path == "/" {
		return nil
	}

	dpath, name := path.Split(info.Path)
	dpath = strings.TrimRight(dpath, "/")

	if dpath == "" {
		dpath = "/"
	}

	dinfo, dir, err := eft.getDir(snap, dpath)
	if err != nil {
		return trace(err)
	}

	dir[name] = info.Type

	dinfo.ModT = info.ModT

	err = eft.putDir(snap, dinfo, dir)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) ListInfos(dpath string) ([]ItemInfo, error) {
	eft.begin()

	_, dir, err := eft.getDir(eft.mainSnap(), dpath)
	if err != nil {
		eft.abort()
		return nil, trace(err)
	}
	
	eft.commit()

	names := make([]string, 0)

	for kk, _ := range(dir) {
		names = append(names, kk)
	}

	sort.Strings(names)

	infos := make([]ItemInfo, 0)

	for _, nn := range(names) {
		info, err := eft.GetInfo(path.Join(dpath, nn))
		if err != nil {
			return infos, trace(err)
		}

		infos = append(infos, info)
	}

	return infos, nil
}

func (eft *EFT) ListDir(dpath string) ([]string, error) {
	infos, err := eft.ListInfos(dpath)
	if err != nil {
		return nil, trace(err)
	}

	items := make([]string, 0)

	for _, info := range(infos) {
		_, name := path.Split(info.Path)
		desc := fmt.Sprintf("%s (%s) %d", name, info.TypeName(), info.Size)
		items = append(items, desc)
	}

	return items, nil
}

func (eft *EFT) listSubInfos(snap *Snapshot, dpath string) ([]*ItemInfo, error) {
	infos := make([]*ItemInfo, 0)

	dinfo, dir, err := eft.getDir(snap, dpath)
	if err != nil {
		eft.abort()
		return nil, trace(err)
	}

	infos = append(infos, &dinfo)

	for kk, _ := range(dir) {
		info, _, err := eft.getTree(snap, path.Join(dpath, kk))
		if err != nil {
			return nil, trace(err)
		}

		if info.Type == INFO_DIR {
			sub_infos, err := eft.listSubInfos(snap, info.Path)
			if err != nil {
				return nil, trace(err)
			}

			infos = append(infos, sub_infos...)
		} else {
			infos = append(infos, &info)
		}
	}

	return infos, nil
}

func (eft *EFT) ListAllInfos() ([]*ItemInfo, error) {
	eft.begin()

	infos, err := eft.listSubInfos(eft.mainSnap(), "/")
	if err != nil {
		eft.abort()
		return nil, trace(err)
	}

	eft.commit()

	paths := make([]string, 0)
	p_map := make(map[string]*ItemInfo)

	for _, info := range(infos) {
		paths = append(paths, info.Path)
		p_map[info.Path] = info
	}

	sort.Strings(paths)

	sorted := make([]*ItemInfo, 0) 

	for _, pp := range(paths) {
		sorted = append(sorted, p_map[pp])
	}

	return sorted, nil
}

