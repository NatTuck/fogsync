package eft

import (
	"fmt"
	"os"
)

type PathTrie struct {
	root *TrieNode
}

func (pt *PathTrie) KeyBytes(ee TrieEntry) ([]byte) {
	info := pt.root.eft.loadItemInfo(ee.Hash)
	hash := HashString(info.Path)
	return hash[:]
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

func (eft *EFT) putTree(info ItemInfo, data_hash [32]byte) ([32]byte, error) {
	trie := eft.emptyPathTrie()

	root_hash, err := eft.getRoot()
	if err == nil {
		trie, err = eft.loadPathTrie(root_hash)
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

func (eft *EFT) getTree(item_path string) (ItemInfo, [32]byte, error) {
	info := ItemInfo{}

	item_hash := [32]byte{}

	root_hash, err := eft.getRoot()
	if err != nil {
		return info, item_hash, ErrNotFound
	}

	trie, err := eft.loadPathTrie(root_hash)
	if err != nil {
		return info, item_hash, trace(err)
	}

	item_hash, err = trie.find(item_path)
	if err != nil {
		return info, item_hash, err // Could be ErrNotFound
	}

	info = eft.loadItemInfo(item_hash)
	if err != nil {
		return info, item_hash, trace(err)
	}

	return info, item_hash, nil
}

func (eft *EFT) delTree(item_path string) [32]byte {
	info0, _, err := eft.getTree(item_path)
	assert_no_error(err)

	info1, err := MakeTombstone(info0)
	assert_no_error(err)

	temp_name := eft.TempName()
	temp, err := os.Create(temp_name)
	assert_no_error(err)
	temp.Close()
	defer os.Remove(temp_name)

	root := eft.putItem(info1, temp_name)
	assert_no_error(err)

	return root
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
	infos := make([]ItemInfo, 0)

	err := eft.with_read_lock(func() {
		root_hash, err := eft.getRoot()
		assert_no_error(err)

		pt, err := eft.loadPathTrie(root_hash)
		assert_no_error(err)
		
		err = pt.root.visitEachEntry(func (ent *TrieEntry) error {
			if ent.Type == TRIE_TYPE_ITEM {
				info := eft.loadItemInfo(ent.Hash)
				infos = append(infos, info)
			}
			
			return nil
		})
		assert_no_error(err)
	})
		
	return infos, err
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
			info := tn.eft.loadItemInfo(ent.Hash)
			fmt.Println(indent(depth), ii, "\tITEM", info.String())

			tn.eft.debugDumpItem(ent.Hash, depth + 1)
		default:
			fmt.Println(indent(depth), ii, "** UKNOWN **", ent.Type)
		}
	}

	fmt.Println(indent(depth), "Skipped empties:", empties)

}
