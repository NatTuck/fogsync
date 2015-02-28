package eft

// Given a set of snapshot root hashes,
// delete all blocks not reachable from those roots.

import (
	"path/filepath"
	"fmt"
	"os"
)

func (eft *EFT) collect() error {
	snaps := eft.listSnaps()

	// Generate full block set.
	live := eft.NewBlockSet()

	for _, name := range(snaps) {
		snap_root, err := eft.getSnapRoot(name)
		assert_no_error(err)

		live.Add(snap_root)

		pt, err := eft.loadPathTrie(snap_root)
		assert_no_error(err)

		bs := pt.blockSet()
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

	err := filepath.Walk(blocks_dir, walk_fn)
	if err != nil {
		return trace(err)
	}

	return nil
}
