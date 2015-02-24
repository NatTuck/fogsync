package eft

// Given a set of snapshot root hashes,
// delete all blocks not reachable from those roots.

import (
	"path/filepath"
	"fmt"
	"os"
)

func (eft *EFT) collect() error {
	snaps, err := eft.loadSnaps()
	if err != nil {
		return trace(err)
	}

	// Generate full block set.
	live := eft.NewBlockSet()

	shash, err := eft.loadSnapsHash()
	if err != nil {
		return trace(err)
	}

	live.Add(shash)

	for _, snap := range(snaps) {
		bs := snap.liveBlocks()
		live.AddSet(bs)
	}

	fmt.Println("Live size:", live.Size())

	// Traverse blocks and delete extras.
	walk_fn := func(pp string, sysi os.FileInfo, err error) error {
		if err != nil {
			return trace(err)
		}

		_, name := filepath.Split(pp)
		if len(name) != 64 {
			return nil
		}

		hash := HexToHash(name)

		if !live.Has(hash) {
			err := os.Remove(pp)
			if err != nil {
				return trace(err)
			}
		}

		return nil
	}

	blocks_dir := filepath.Join(eft.Dir, "blocks")

	err = filepath.Walk(blocks_dir, walk_fn)
	if err != nil {
		return trace(err)
	}

	return nil
}
