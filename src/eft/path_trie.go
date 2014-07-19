package eft

import (
	"fmt"
)

type PathTrie struct {
	root *TrieNode
}

func (eft *EFT) emptyPathTrie() PathTrie {
	getPathKey := func(ee TrieEntry) ([]byte, error) {
		info, err := eft.loadItemInfo(ee.Hash[:])
		if err != nil {
			return nil, err
		}

		return HashString(info.Path), nil
	}

	trie := PathTrie{}

	trie.root = &TrieNode{
		eft: eft,
		key: getPathKey,
	}

	return trie
}

func (eft *EFT) loadPathTrie(hash []byte) (PathTrie, error) {
	trie := eft.emptyPathTrie()

	err := trie.root.load(hash)
	if err != nil {
		return trie, trace(err)
	}

	return trie, nil
}

func (pt *PathTrie) save() ([]byte, error) {
	return pt.root.save()
}

func (pt *PathTrie) find(item_path string) ([]byte, error) {
	path_hash := HashString(item_path)
	return pt.root.find(path_hash, 0)
}

func (pt *PathTrie) insert(item_path string, data_hash []byte) error {
	path_hash := HashString(item_path)

	entry := TrieEntry{}
	copy(entry.Hash[:], data_hash)

	return pt.root.insert(path_hash, entry, 0)
}

func (pt *PathTrie) remove(name string) error {
	path_hash := HashString(name)
	return pt.root.remove(path_hash, 0)
}

func (eft *EFT) putTree(snap *Snapshot, info ItemInfo, data_hash []byte) ([]byte, error) {
	trie := eft.emptyPathTrie()

	var err error

	if !snap.isEmpty() {
		trie, err = eft.loadPathTrie(snap.Root[:])
		if err != nil {
			return nil, trace(err)
		}
	}

	err = trie.insert(info.Path, data_hash)
	if err != nil {
		return nil, trace(err)
	}

	root_hash, err := trie.save()
	if err != nil {
		return nil, trace(err)
	}

	return root_hash, nil
}

func (eft *EFT) getTree(snap *Snapshot, item_path string) (ItemInfo, []byte, error) {
	info := ItemInfo{}

	if snap.isEmpty() {
		return info, nil, ErrNotFound 
	}

	trie, err := eft.loadPathTrie(snap.Root[:])
	if err != nil {
		return info, nil, trace(err)
	}

	item_hash, err := trie.find(item_path)
	if err != nil {
		return info, nil, err // Could be ErrNotFound
	}

	info, err = eft.loadItemInfo(item_hash)
	if err != nil {
		return info, nil, err
	}

	return info, item_hash, nil
}

func (eft *EFT) delTree(snap *Snapshot, item_path string) ([]byte, error) {
	trie, err := eft.loadPathTrie(snap.Root[:])
	if err != nil {
		return nil, trace(err)
	}
	
	err = trie.remove(item_path)
	if err != nil {
		return nil, trace(err)
	}

	root_hash, err := trie.save()
	if err != nil {
		return nil, trace(err)
	}

	return root_hash, nil
}

func (pt *PathTrie) visitEachBlock(fn func(hash []byte) error) error {
	return pt.root.visitEachEntry(func (ent *TrieEntry) error {
		switch ent.Type {
		case TRIE_TYPE_MORE:
			return fn(ent.Hash[:])

		case TRIE_TYPE_ITEM:
			err := fn(ent.Hash[:])
			if err != nil {
				return trace(err)
			}

			eft := pt.root.eft
		
			info, err := eft.loadItemInfo(ent.Hash[:])
			if err != nil {
				return trace(err)
			}

			err = eft.visitItemBlocks(info, fn)
			if err != nil {
				return trace(err)
			}

			return nil

		default:
			panic(fmt.Sprintf("Can't handle entry of type %d\n", ent.Type))
		}
	})
}
