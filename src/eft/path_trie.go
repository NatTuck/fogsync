package eft

import (
	"fmt"
)

type PathTrie struct {
	root *TrieNode
}

func (eft *EFT) emptyPathTrie() PathTrie {
	getPathKey := func(ee TrieEntry) ([]byte, error) {
		info, err := eft.loadItemInfo(ee.Hash)
		if err != nil {
			return nil, err
		}

		hash := HashString(info.Path)
		return hash[:], nil
	}

	trie := PathTrie{}

	trie.root = &TrieNode{
		eft: eft,
		key: getPathKey,
	}

	return trie
}

func (eft *EFT) loadPathTrie(hash [32]byte) (PathTrie, error) {
	trie := eft.emptyPathTrie()

	err := trie.root.load(hash)
	if err != nil {
		return trie, trace(err)
	}

	return trie, nil
}

func (pt *PathTrie) save() ([32]byte, error) {
	return pt.root.save()
}

func (pt *PathTrie) find(item_path string) ([32]byte, error) {
	path_hash := HashString(item_path)
	return pt.root.find(path_hash[:], 0)
}

func (pt *PathTrie) insert(item_path string, data_hash [32]byte) error {
	path_hash := HashString(item_path)

	entry := TrieEntry{}
	entry.Hash = data_hash

	return pt.root.insert(path_hash[:], entry, 0)
}

func (pt *PathTrie) remove(name string) error {
	path_hash := HashString(name)
	return pt.root.remove(path_hash[:], 0)
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
		return info, item_hash, err
	}

	return info, item_hash, nil
}

func (eft *EFT) delTree(snap *Snapshot, item_path string) ([32]byte, error) {
	empty := [32]byte{}

	trie, err := eft.loadPathTrie(snap.Root)
	if err != nil {
		return empty, trace(err)
	}
	
	err = trie.remove(item_path)
	if err != nil {
		return empty, trace(err)
	}

	root_hash, err := trie.save()
	if err != nil {
		return empty, trace(err)
	}

	return root_hash, nil
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
