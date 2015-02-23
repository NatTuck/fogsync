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

func (pt *PathTrie) Equals(pt1 *PathTrie) bool {
	return pt.root.Equals(pt1.root)
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

func (snap *Snapshot) putTree(info ItemInfo, data_hash [32]byte) ([32]byte, error) {
	trie := snap.eft.emptyPathTrie()

	var err error
	root_hash := [32]byte{}

	if !snap.isEmpty() {
		trie, err = snap.eft.loadPathTrie(snap.Root)
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

func (snap *Snapshot) getTree(item_path string) (ItemInfo, [32]byte, error) {
	info := ItemInfo{}

	item_hash := [32]byte{}

	if snap.isEmpty() {
		return info, item_hash, ErrNotFound 
	}

	trie, err := snap.eft.loadPathTrie(snap.Root)
	if err != nil {
		return info, item_hash, trace(err)
	}

	item_hash, err = trie.find(item_path)
	if err != nil {
		return info, item_hash, err // Could be ErrNotFound
	}

	info, err = snap.eft.loadItemInfo(item_hash)
	if err != nil {
		return info, item_hash, trace(err)
	}

	/*
	if info.Type == INFO_TOMB {
		return info, item_hash, ErrNotFound
	}
	*/

	return info, item_hash, nil
}

func (snap *Snapshot) delTree(item_path string) ([32]byte, error) {
	empty := [32]byte{}

	info0, _, err := snap.getTree(item_path)
	if err != nil {
		return empty, trace(err)
	}

	info1, err := MakeTombstone(info0)
	if err != nil {
		return empty, trace(err)
	}

	temp_name := snap.eft.TempName()
	temp, err := os.Create(temp_name)
	if err != nil {
		return empty, trace(err)
	}
	temp.Close()
	defer os.Remove(temp_name)

	err = snap.putItem(info1, temp_name)
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

func (pt *PathTrie) blockSet() *BlockSet {
	bs := pt.root.eft.NewBlockSet()

	err := pt.visitEachBlock(func (hash [32]byte) error {
		bs.Add(hash)
		return nil
	})
	assert_no_error(err)

	return bs
}

func (eft *EFT) ListInfos() ([]ItemInfo, error) {
	snap, err := eft.GetSnap("")
	if err != nil {
		return nil, trace(err)
	}

	return snap.ListInfos()
}

func (snap *Snapshot) ListInfos() ([]ItemInfo, error) {
	pt, err := snap.eft.loadPathTrie(snap.Root)
	if err != nil {
		return nil, trace(err)
	}

	infos := make([]ItemInfo, 0)

	err = pt.root.visitEachEntry(func (ent *TrieEntry) error {
		if ent.Type == TRIE_TYPE_ITEM {
			info, err := snap.eft.loadItemInfo(ent.Hash)
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

func (pt *PathTrie) debugDump(depth int) {
	fmt.Println(indent(depth), "[PathTrie]")
	tn := pt.root

	fmt.Println(indent(depth), "[TrieNode]")

	empties := 0

	for ii := 0; ii < 256; ii++ {
		ent := &tn.tab[ii]

		switch ent.Type {
		case TRIE_TYPE_NONE:
			empties += 1
		case TRIE_TYPE_MORE:
			fmt.Println(indent(depth), ii, "\tMORE");
			child, err := tn.eft.loadPathTrie(ent.Hash)
			if err != nil {
				panic(err)
			}
			child.debugDump(depth + 1)
		case TRIE_TYPE_OVRF:
			fmt.Println(indent(depth), ii, "\tOVRF")
		case TRIE_TYPE_ITEM:
			info, err := tn.eft.loadItemInfo(ent.Hash)
			if err != nil {
				panic(err)
			}
			fmt.Println(indent(depth), ii, "\tITEM", info.String())

			tn.eft.debugDumpItem(ent.Hash, depth + 1)
		default:
			fmt.Println(indent(depth), ii, "** UKNOWN **", ent.Type)
		}
	}

	fmt.Println(indent(depth), "Skipped empties:", empties)

}
