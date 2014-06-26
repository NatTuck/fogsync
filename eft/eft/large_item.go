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

func getLargeKey(ee TrieEntry) ([]byte, error) {
	return ee.Pkey[:], nil
}

func (eft *EFT) newLargeTrie(info ItemInfo) LargeTrie {
	trie := LargeTrie{
		info: info,
	}

	trie.root = &TrieNode{
		eft: eft,
		key: getLargeKey,
	}

	return trie
}

func (eft *EFT) loadLargeTrie(hash []byte) (LargeTrie, error) {
	trie := LargeTrie{}

	trie.root = &TrieNode{
		eft: eft,
		key: getLargeKey,
	}
	
	err := trie.root.load(hash)
	if err != nil {
		return trie, trace(err)
	}

	trie.info = ItemInfoFromBytes(trie.root.hdr[:])

	return trie, nil
}

func (trie *LargeTrie) save() ([]byte, error) {
	copy(trie.root.hdr[:], trie.info.Bytes())
	return trie.root.save()
}

func (trie *LargeTrie) find(ii uint64) ([]byte, error) {
	le := binary.LittleEndian

	var iile [8]byte
	le.PutUint64(iile[:], ii)

	return trie.root.find(iile[:], 0)
}

func (trie *LargeTrie) insert(ii uint64, hash []byte) error {
	le := binary.LittleEndian

	entry := TrieEntry{}
	copy(entry.Hash[:], hash)
	le.PutUint64(entry.Pkey[:], ii)

	return trie.root.insert(entry.Pkey[:], entry, 0)
}

func (eft *EFT) saveLargeItem(info ItemInfo, src_path string) ([]byte, error) {
	src, err := os.Open(src_path)
	if err != nil {
		return nil, trace(err)
	}
	defer src.Close()

	trie := eft.newLargeTrie(info)

	data := make([]byte, BLOCK_SIZE)

	for ii := uint64(0); true; ii++ {
		_, err := src.Read(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, trace(err)
		}

		b_hash, err := eft.saveBlock(data)
		if err != nil {
			return nil, trace(err)
		}

		err = trie.insert(ii, b_hash)
		if err != nil {
			return nil, trace(err)
		}

		// XX - Remove this nonsense
		hash, err := trie.find(ii)
		if err != nil {
			fmt.Println("XX - Insert failed", ii)
			panic(err)
		}

		if !BytesEqual(hash, b_hash) {
			fmt.Println("XX - Insert failed", ii)
			panic("Argh!")
		}
	}

	hash, err := trie.save()
	if err != nil {
		return nil, trace(err)
	}

	return hash, nil
}

func (eft *EFT) loadLargeItem(hash []byte, dst_path string) (_ ItemInfo, eret error) {
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

func (eft *EFT) killLargeItemBlocks(hash []byte) error {
	trie, err := eft.loadLargeTrie(hash)
	if err != nil {
		return trace(err)
	}

	for ii := uint64(0); true; ii++ {
		b_hash, err := trie.find(ii)
		if err == ErrNotFound {
			break
		}
		if err != nil {
			return trace(err)
		}

		err = eft.pushDead(b_hash)
		if err != nil {
			return trace(err)
		}
	}

	return nil
}

