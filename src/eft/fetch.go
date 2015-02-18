package eft

import (
	"fmt"
)

type FetchFn func(bs *BlockSet) (*BlockArchive, error)

func (eft *EFT) FetchRemote(rem_hash [32]byte, fetch_fn FetchFn) error {
	eft.Lock()
	defer eft.Unlock()

	bs, err := eft.NewBlockSet1(rem_hash)
	if err != nil {
		return trace(err)
	}

	err = eft.fetchBlocks(bs, fetch_fn)
	if err != nil {
		return trace(err)
	}

	snaps, err := eft.loadSnapsFrom(rem_hash)
	if err != nil {
		return trace(err)
	}

	for _, snap := range(snaps) {
		err := eft.fetchSnap(snap, fetch_fn)
		if err != nil {
			return trace(err)
		}
	}

	return nil
}

func (eft *EFT) fetchSnap(snap *Snapshot, fetch_fn FetchFn) error {
	bs, err := eft.NewBlockSet1(snap.Root)
	if err != nil {
		return trace(err)
	}

	err = eft.fetchBlocks(bs, fetch_fn)
	if err != nil {
		return trace(err)
	}

	pt, err := eft.loadPathTrie(snap.Root)
	if err != nil {
		return trace(err)
	}

	err = eft.fetchPathTrieNode(pt.root, 0, fetch_fn)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) fetchPathTrieNode(ptn *TrieNode, dd int, fetch_fn FetchFn) error {
	// First, fetch all sub-blocks
	bs, err := eft.NewBlockSet()
	if err != nil {
		return trace(err)
	}

	for _, ent := range(ptn.tab) {
		if ent.Type != TRIE_TYPE_NONE {
			err := bs.Add(ent.Hash)
			if err != nil {
				return trace(err)
			}
		}
	}

	err = eft.fetchBlocks(bs, fetch_fn)
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
	info, err := eft.loadItemInfo(hash)
	if err != nil {
		return trace(err)
	}

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
	bs, err := eft.NewBlockSet()
	if err != nil {
		return trace(err)
	}

	for _, ent := range(ltn.tab) {
		if ent.Type != TRIE_TYPE_NONE {
			err := bs.Add(ent.Hash)
			if err != nil {
				return trace(err)
			}
		}
	}
	
	err = eft.fetchBlocks(bs, fetch_fn)
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
	ba, err := fetch_fn(bs)
	if err != nil {
		return trace(err)
	}

	err = ba.Extract(eft)
	if err != nil {
		return trace(err)
	}

	return nil
}


