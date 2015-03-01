package eft

import (
	"os"
	"io"
	"fmt"
	"encoding/binary"
)

type LargeTrie struct {
	info  ItemInfo
	root *TrieNode
}

func (trie *LargeTrie) KeyBytes(ee TrieEntry) []byte {
	return ee.Pkey[:]
}

func (eft *EFT) newLargeTrie(info ItemInfo) LargeTrie {
	trie := LargeTrie{
		info: info,
	}

	trie.root = &TrieNode{
		eft: eft,
		tri: &trie,
		dep: 0,
	}

	return trie
}

func (eft *EFT) loadLargeTrie(hash [32]byte) (LargeTrie, error) {
	trie := LargeTrie{}

	trie.root = &TrieNode{
		eft: eft,
		tri: &trie,
		dep: 0,
	}
	
	err := trie.root.load(hash)
	if err != nil {
		return trie, trace(err)
	}

	trie.info = ItemInfoFromBytes(trie.root.hdr[:])

	return trie, nil
}

func (trie *LargeTrie) save() ([32]byte, error) {
	copy(trie.root.hdr[:], trie.info.Bytes())
	return trie.root.save()
}

func (trie *LargeTrie) find(ii uint64) ([32]byte, error) {
	le := binary.LittleEndian

	var iile [8]byte
	le.PutUint64(iile[:], ii)
	
	return trie.root.find(iile[:])
}

func (trie *LargeTrie) insert(ii uint64, hash [32]byte) error {
	le := binary.LittleEndian

	entry := TrieEntry{}
	entry.Hash = hash
	le.PutUint64(entry.Pkey[:], ii)

	return trie.root.insert(entry.Pkey[:], entry)
}

func (eft *EFT) saveLargeItem(info ItemInfo, src_path string) ([32]byte, error) {
	hash := [32]byte{}

	src, err := os.Open(src_path)
	if err != nil {
		return hash, trace(err)
	}
	defer src.Close()

	trie := eft.newLargeTrie(info)

	data := make([]byte, DATA_SIZE)

	for ii := uint64(0); true; ii++ {
		_, err := src.Read(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			return hash, trace(err)
		}

		b_hash, err := eft.saveBlock(data)
		if err != nil {
			return hash, trace(err)
		}

		err = trie.insert(ii, b_hash)
		if err != nil {
			return hash, trace(err)
		}
	}

	hash, err = trie.save()
	if err != nil {
		return hash, trace(err)
	}

	return hash, nil
}

func (eft *EFT) loadLargeItem(hash [32]byte, dst_path string) (_ ItemInfo, eret error) {
	info := ItemInfo{}

	dst, err := os.Create(dst_path)
	if err != nil {
		return info, trace(err)
	}
	defer func() {
		eret = dst.Close()
	}()

	trie, err := eft.loadLargeTrie(hash)
	if err != nil {
		return info, trace(err)
	}

	info = trie.info

	for ii := uint64(0); true; ii++ {
		b_hash, err := trie.find(ii)
		if err == ErrNotFound {
			break
		}
		if err != nil {
			return info, trace(err)
		}

		data, err := eft.loadBlock(b_hash)
		if err != nil {
			return info, trace(err)
		}

		_, err = dst.Write(data)
		if err != nil {
			return info, trace(err)
		}
	}

	sysi, err := os.Lstat(dst_path)
	if err != nil {
		return info, trace(err)
	}

	if uint64(sysi.Size()) < info.Size {
		return info, trace(fmt.Errorf("Extracted item too small"))
	}

	err = dst.Truncate(int64(info.Size))
	if err != nil {
		return info, trace(err)
	}

	return info, nil
}

func (lt *LargeTrie) visitEachBlock(fn func(hash [32]byte) error) error {
	return lt.root.visitEachEntry(func(ent *TrieEntry) error {
		return fn(ent.Hash)
	})
}

func (lt *LargeTrie) debugDump(deep int) {
	fmt.Println(indent(deep), "[LargeTrie]")
	tn := lt.root

	fmt.Println(indent(deep), "[TrieNode]")

	empties := 0

	for ii := 0; ii < 256; ii++ {
		ent := &tn.tab[ii]

		switch ent.Type {
		case TRIE_TYPE_NONE:
			empties += 1
		case TRIE_TYPE_MORE:
			fmt.Printf("%s %4d \tMORE\n", indent(deep), ii);
			child, err := tn.eft.loadLargeTrie(ent.Hash)
			if err != nil {
				panic(err)
			}
			child.debugDump(deep + 1)
		case TRIE_TYPE_OVRF:
			fmt.Printf("%s %4d \tOVRF\n", indent(deep), ii);
		case TRIE_TYPE_ITEM:
			fmt.Printf("%s %4d \tITEM\n", indent(deep), ii);
		default:
			fmt.Println(indent(deep), ii, "** UKNOWN **", ent.Type)
		}
	}

	fmt.Println(indent(deep), "Skipped empties:", empties)

}
