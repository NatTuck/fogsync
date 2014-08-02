package eft

import (
	"fmt"
	"os"
)

type PathTrie struct {
	root *TrieNode
}

func (pt *PathTrie) KeyBytes(ee TrieEntry) ([]byte, error) {
	info, err := pt.root.eft.loadItemInfo(ee.Hash)
	if err != nil {
		return nil, err
	}

	hash := HashString(info.Path)
	return hash[:], nil
}

func (eft *EFT) emptyPathTrie() PathTrie {
	trie := PathTrie{}

	trie.root = &TrieNode{
		eft: eft,
		tri: &trie,
		dep: 0,
	}

	return trie
}

func (eft *EFT) loadPathTrie(hash [32]byte) (PathTrie, error) {
	trie := eft.emptyPathTrie()

	if hash != ZERO_HASH {
		err := trie.root.load(hash)
		if err != nil {
			return trie, trace(err)
		}
	}

	return trie, nil
}

func (pt *PathTrie) save() ([32]byte, error) {
	return pt.root.save()
}

func (pt *PathTrie) find(item_path string) ([32]byte, error) {
	path_hash := HashString(item_path)
	return pt.root.find(path_hash[:])
}

func (pt *PathTrie) insert(item_path string, data_hash [32]byte) error {
	path_hash := HashString(item_path)

	entry := TrieEntry{}
	entry.Hash = data_hash

	return pt.root.insert(path_hash[:], entry)
}

func (eft *EFT) putTree(snap *Snapshot, info ItemInfo, data_hash [32]byte) ([32]byte, error) {
	trie := eft.emptyPathTrie()

	var err error
	root_hash := [32]byte{}

	if !snap.isEmpty() {
		trie, err = eft.loadPathTrie(snap.Root)
		if err != nil {
			return root_hash, trace(err)
		}
	}

	err = trie.insert(info.Path, data_hash)
	if err != nil {
		return root_hash, trace(err)
	}

	root_hash, err = trie.save()
	if err != nil {
		return root_hash, trace(err)
	}

	return root_hash, nil
}

func (eft *EFT) getTree(snap *Snapshot, item_path string) (ItemInfo, [32]byte, error) {
	info := ItemInfo{}

	item_hash := [32]byte{}

	if snap.isEmpty() {
		return info, item_hash, ErrNotFound 
	}

	trie, err := eft.loadPathTrie(snap.Root)
	if err != nil {
		return info, item_hash, trace(err)
	}

	item_hash, err = trie.find(item_path)
	if err != nil {
		return info, item_hash, err // Could be ErrNotFound
	}

	info, err = eft.loadItemInfo(item_hash)
	if err != nil {
		return info, item_hash, trace(err)
	}

	if info.Type == INFO_TOMB {
		return info, item_hash, ErrNotFound
	}

	return info, item_hash, nil
}

func (eft *EFT) delTree(snap *Snapshot, item_path string) ([32]byte, error) {
	empty := [32]byte{}

	info0, _, err := eft.getTree(snap, item_path)
	if err != nil {
		return empty, trace(err)
	}

	info1, err := MakeTombstone(info0)
	if err != nil {
		return empty, trace(err)
	}

	temp_name := eft.TempName()
	temp, err := os.Create(temp_name)
	if err != nil {
		return empty, trace(err)
	}
	temp.Close()
	defer os.Remove(temp_name)

	err = eft.putItem(snap, info1, temp_name)
	if err != nil {
		return empty, trace(err)
	}

	return snap.Root, nil
}

func (pt *PathTrie) visitEachBlock(fn func(hash [32]byte) error) error {
	return pt.root.visitEachEntry(func (ent *TrieEntry) error {
		switch ent.Type {
		case TRIE_TYPE_MORE:
			return fn(ent.Hash)

		case TRIE_TYPE_ITEM:
			eft := pt.root.eft
		
			err := eft.visitItemBlocks(ent.Hash, fn)
			if err != nil {
				return trace(err)
			}

			return nil

		default:
			panic(fmt.Sprintf("Can't handle entry of type %d\n", ent.Type))
		}
	})
}

func (eft *EFT) ListInfos() ([]ItemInfo, error) {
	eft.Lock()
	defer eft.Unlock()

	snap := eft.mainSnap()

	pt, err := eft.loadPathTrie(snap.Root)
	if err != nil {
		return nil, trace(err)
	}

	infos := make([]ItemInfo, 0)

	err = pt.root.visitEachEntry(func (ent *TrieEntry) error {
		if ent.Type == TRIE_TYPE_ITEM {
			info, err := eft.loadItemInfo(ent.Hash)
			if err != nil {
				return trace(err)
			}

			infos = append(infos, info)
		}

		return nil
	})
	if err != nil {
		return nil, trace(err)
	}

	return infos, nil
}
