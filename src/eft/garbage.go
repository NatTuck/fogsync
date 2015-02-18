package eft

// Unused blocks are removed using a mark and sweep technique
// as follows:
//
//  - A list of all blocks is generated by scanning the file system.
//  - That list is sorted.
//  - The EFT is traversed, and all used blocks are marked.
//  - Unmarked blocks are deleted, and a list of unmarked blocks is
//    saved to apply to the cloud server.

import (
	"github.com/edsrzf/mmap-go"
	"encoding/hex"
	"sort"
	"bytes"
	"path/filepath"
	"os"
	"fmt"
)

func (eft *EFT) collect() (_ string, eret error) {
	mm, err := eft.newMarkList()
	if err != nil {
		return "", trace(err)
	}
	defer func() {
		err := mm.Close()
		if eret == nil && err != nil {
			eret = trace(err)
		}
	}()

	err = mm.scan()
	if err != nil {
		return "", trace(err)
	}

	err = mm.mark()
	if err != nil {
		return "", trace(err)
	}

	dead_name, err := mm.sweep()
	if err != nil {
		return "", trace(err)
	}

	return dead_name, nil
}

type MarkList struct {
	eft  *EFT
	name string
	file *os.File
	data mmap.MMap
	size int
}

func (eft *EFT) newMarkList() (*MarkList, error) {
	mm := &MarkList{ eft: eft }
	mm.name = eft.TempName()

	return mm, nil
}

func (mm *MarkList) put(ii int, hash [32]byte, marked bool) {
	pos := ii * 33 

	copy(mm.data[pos:pos + 32], hash[:])

	if marked {
		mm.data[pos + 32] = 1
	}
}

func (mm *MarkList) get(ii int) ([32]byte, bool) {
	pos := ii * 33

	hash := [32]byte{}
	copy(hash[:], mm.data[pos:pos + 32])

	if mm.data[pos + 32] > 0 {
		return hash, true
	} else {
		return hash, false
	}
}

func (mm *MarkList) Len() int {
	return mm.size
}

func (mm *MarkList) Less(ii, jj int) bool {
	aa, _ := mm.get(ii)
	bb, _ := mm.get(jj)
	return bytes.Compare(aa[:], bb[:]) < 0
}

func (mm *MarkList) Swap(ii, jj int) {
	aa, aam := mm.get(ii)
	bb, bbm := mm.get(jj)
	mm.put(ii, bb, bbm)
	mm.put(jj, aa, aam)
}

func (mm *MarkList) openFile() error {
	flags := os.O_CREATE | os.O_RDWR
	file, err := os.OpenFile(mm.name, flags, 0600)
	if err != nil {
		return trace(err)
	}

	mm.file = file
	return nil
}

func (mm *MarkList) scan() error {
	err := mm.openFile()
	if err != nil {
		return trace(err)
	}

	walk_fn := func(pp string, sysi os.FileInfo, err error) error {
		if err != nil {
			return trace(err)
		}

		_, name := filepath.Split(pp)
		if len(name) != 64 {
			return nil
		}

		hash, err := hex.DecodeString(name)
		if err != nil {
			return trace(err)
		}

		item := make([]byte, 33)
		copy(item[0:32], hash)

		_, err = mm.file.Write(item)
		if err != nil {
			return trace(err)
		}

		mm.size = mm.size + 1

		return nil
	}

	blocks_dir := filepath.Join(mm.eft.Dir, "blocks")

	err = filepath.Walk(blocks_dir, walk_fn)
	if err != nil {
		return trace(err)
	}

	sysi, err := os.Lstat(mm.name)
	if err != nil {
		return trace(err)
	}

	if sysi.Size() == 0 {
		return trace(fmt.Errorf("No blocks"))
	}

	data, err := mmap.Map(mm.file, mmap.RDWR, 0)
	if err != nil {
		return trace(err)
	}

	mm.data = data

	return nil
}

func (mm *MarkList) mark() error {
	// To mark, sort the list of blocks then traverse the EFT
	// and mark all used blocks.
	sort.Sort(mm)

	hash, err := mm.eft.loadSnapsHash()
	if err != nil {
		return trace(err)
	}

	err = mm.markBlock(hash)
	if err != nil {
		return trace(err)
	}

	snaps, err := mm.eft.loadSnaps()
	if err != nil {
		return trace(err)
	}

	for _, snap := range(snaps) {
		err := mm.markPathTrie(snap.Root)
		if err != nil {
			return trace(err)
		}
	}

	return nil
}

func (mm *MarkList) markBlock(hash [32]byte) error {
	// Find entry
	ii := sort.Search(mm.Len(), func (jj int) bool {
		aa, _ := mm.get(jj)
		return bytes.Compare(aa[:], hash[:]) >= 0
	})

	aa, _ := mm.get(ii)
	if !HashesEqual(aa, hash) {
		text := hex.EncodeToString(hash[:])
		return trace(fmt.Errorf("Attempted to mark non-existant block: %s", text))
	}

	mm.put(ii, aa, true)
	return nil
}

func (mm *MarkList) markPathTrie(hash [32]byte) error {
	err := mm.markBlock(hash)
	if err != nil {
		return trace(err)
	}

	trie, err := mm.eft.loadPathTrie(hash)
	if err != nil {
		return trace(err)
	}

	err = trie.visitEachBlock(func (hash [32]byte) error {
		return mm.markBlock(hash)
	})
	if err != nil {
		return trace(err)
	}

	return nil
}

func (mm *MarkList) markItem(hash [32]byte) error {
	err := mm.eft.visitItemBlocks(hash, func(hash [32]byte) error {
		return mm.markBlock(hash)
	})
	if err != nil {
		return trace(err)
	}

	return nil
}

func (mm *MarkList) sweep() (string, error) {
	dead_name := mm.eft.TempName()
	dead, err := os.Create(dead_name)
	if err != nil {
		return "", trace(err)
	}

	for ii := 0; ii < mm.Len(); ii++ {
		hash, mark := mm.get(ii)
		if !mark {
			text := hex.EncodeToString(hash[:])
			_, err = dead.Write([]byte(text + "\n"))
			if err != nil {
				return "", trace(err)
			}
		}
	}

	// Remove Blocks

	return dead_name, nil
}

func (mm *MarkList) Close() error {
	err := mm.data.Unmap()
	if err != nil {
		return trace(err)
	}

	err = mm.file.Close()
	if err != nil {
		return trace(err)
	}

	return os.Remove(mm.name)
}

func (snap *Snapshot) liveBlocks() ([]string, error) {
	trie, err := snap.eft.loadPathTrie(snap.Root)
	if err != nil {
		return nil, trace(err)
	}

	err = trie.visitEachBlock(func (hash [32]byte) error {
		return nil
	})
	if err != nil {
		return nil, trace(err)
	}

	return make([]string, 0), nil
}
