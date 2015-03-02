package eft

import (
	"fmt"
)

type FetchFn func(bs []string) (*BlockArchive, error)

func (eft *EFT) FetchRemote(rem_hash [32]byte, fetch_fn FetchFn) error {
	err := eft.with_read_lock(func() {
		bs := eft.NewBlockSet()
		bs.Add(rem_hash)

		err := eft.fetchBlocks(bs, fetch_fn)
		assert_no_error(err)

		pt, err := eft.loadPathTrie(rem_hash)
		assert_no_error(err)

		err = eft.fetchPathTrieNode(pt.root, 0, fetch_fn)
		assert_no_error(err)

		eft.saveRoot(rem_hash)
	})

	if err != nil {
		return trace(err)
	}

	return eft.with_write_lock(func() {
		// Verify that this root is still there.
		// Could have been lost in GC between locks.
		_, err := eft.loadBlock(rem_hash)
		assert_no_error(err)

		eft.putSnapRoot("remote", rem_hash)
	})
}

func (eft *EFT) fetchPathTrieNode(ptn *TrieNode, dd int, fetch_fn FetchFn) error {
	// First, fetch all sub-blocks
	bs := eft.NewBlockSet()

	for _, ent := range(ptn.tab) {
		if ent.Type != TRIE_TYPE_NONE {
			bs.Add(ent.Hash)
		}
	}

	err := eft.fetchBlocks(bs, fetch_fn)
	if err != nil {
		return trace(err)
	}

	// Next, recurse
	for _, ent := range(ptn.tab) {
		switch ent.Type {
		case TRIE_TYPE_NONE:
			continue

		case TRIE_TYPE_MORE:
			next_ptn, err := ptn.loadChild(ent.Hash)
			if err != nil {
				return trace(err)
			}

			err = eft.fetchPathTrieNode(next_ptn, dd + 1, fetch_fn)
			if err != nil {
				return trace(err)
			}

			continue

		case TRIE_TYPE_OVRF:
			panic("TODO")

		case TRIE_TYPE_ITEM:
			err = eft.fetchItem(ent.Hash, fetch_fn)
			if err != nil {
				return trace(err)
			}

			continue

		default:
			panic(fmt.Sprintf("Unknown trie entry type: %d", ent.Type))
		}
	}

	return nil
}

func (eft *EFT) fetchItem(hash [32]byte, fetch_fn FetchFn) error {
	info := eft.loadItemInfo(hash)

	if (info.Size > 12 * 1024) {
		lgt, err := eft.loadLargeTrie(hash)
		if err != nil {
			return trace(err)
		}

		err = eft.fetchLargeNode(lgt.root, 0, fetch_fn)
		if err != nil {
			return trace(err)
		}
	}

	return nil
}

func (eft *EFT) fetchLargeNode(ltn *TrieNode, dd int, fetch_fn FetchFn) error {
	// First, fetch all the blocks
	bs := eft.NewBlockSet()

	for _, ent := range(ltn.tab) {
		if ent.Type != TRIE_TYPE_NONE {
			bs.Add(ent.Hash)
		}
	}
	
	err := eft.fetchBlocks(bs, fetch_fn)
	if err != nil {
		return trace(err)
	}

	// Then, recurse
	for _, ent := range(ltn.tab) {
		if ent.Type == TRIE_TYPE_MORE {
			next_ltn, err := ltn.loadChild(ent.Hash)
			if err != nil {
				return trace(err)
			}

			err = eft.fetchLargeNode(next_ltn, dd + 1, fetch_fn)
			if err != nil {
				return trace(err)
			}
		}
	}

	return nil
}

func (eft *EFT) fetchBlocks(bs *BlockSet, fetch_fn FetchFn) error {
	ba, err := fetch_fn(bs.HexSlice())
	if err != nil {
		return trace(err)
	}

	err = ba.Extract(eft)
	if err != nil {
		return trace(err)
	}

	return nil
}


