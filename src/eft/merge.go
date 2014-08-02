package eft

import (
	"fmt"
	"bytes"
)

func (eft *EFT) MergeRemote(hash [32]byte) error {
	eft.Lock()
	defer eft.Unlock()

	eft.begin()
	abort := true
	defer func() {
		if abort {
			eft.abort()
		} else {
			eft.commit()
		}
	}()

	// Merge snapshots
	snaps, err := eft.loadSnaps()
	if err != nil {
		return trace(err)
	}

	rem_snaps, err := eft.loadSnapsFrom(hash)
	if err != nil {
		return trace(err)
	}

	if len(snaps) != 1 || len(rem_snaps) != 1 {
		panic("TODO: Handle multiple snapshots")
	}

	merged, err := eft.mergeSnaps(snaps[0], rem_snaps[0])
	if err != nil {
		return trace(err)
	}
	snaps[0] = merged

	err = eft.saveSnaps(snaps)
	if err != nil {
		return trace(err)
	}

	abort = false
	return nil
}

func (eft *EFT) mergeSnaps(snap0, snap1 Snapshot) (Snapshot, error) {
	fmt.Println("XX - Merge Snaps", HashToHex(snap0.Root), HashToHex(snap1.Root))

	if HashesEqual(snap0.Root, snap1.Root) {
		return snap0, nil
	}

	pt0, err := eft.loadPathTrie(snap0.Root)
	if err != nil {
		return Snapshot{}, trace(err)
	}
	
	pt1, err := eft.loadPathTrie(snap1.Root)
	if err != nil {
		return Snapshot{}, trace(err)
	}

	trie, err := eft.mergePathTries(pt0, pt1)
	if err != nil {
		return Snapshot{}, trace(err)
	}

	hash, err := trie.save()
	if err != nil {
		return Snapshot{}, trace(err)
	}

	snap0.Root = hash
	
	fmt.Println("XX - Merged To", HashToHex(hash))

	return snap0, nil
}

func (eft *EFT) mergePathTries(pt0, pt1 PathTrie) (PathTrie, error) {

	mtn, err := eft.mergeTrieNodes(*pt0.root, *pt1.root)
	if err != nil {
		return PathTrie{}, trace(err)
	}

	return PathTrie{root: &mtn}, nil
}

func (eft *EFT) mergeTrieNodes(tn0, tn1 TrieNode) (TrieNode, error) {
	mtn := TrieNode{
		eft: eft,
		dep: tn0.dep,
	}

	for ii := 0; ii < 256; ii++ {
		ent0 := tn0.tab[ii]
		ent1 := tn1.tab[ii]

		if ent0 == ent1 {
			// Same block hash means same entry
			mtn.tab[ii] = ent0
			continue
		}

		if ent0.Type == TRIE_TYPE_NONE {
			mtn.tab[ii] = ent1
			continue
		}

		if ent1.Type == TRIE_TYPE_NONE {
			mtn.tab[ii] = ent0
			continue
		}

		if ent0.Type == TRIE_TYPE_MORE && ent1.Type == TRIE_TYPE_MORE {
			stn0, err := tn0.loadChild(ent0.Hash)
			if err != nil {
				return mtn, trace(err)
			}

			stn1, err := tn1.loadChild(ent1.Hash)
			if err != nil {
				return mtn, trace(err)
			}

			smtn, err := eft.mergeTrieNodes(*stn0, *stn1)
			if err != nil {
				return mtn, trace(err)
			}

			next_hash, err := smtn.save()
			if err != nil {
				return mtn, trace(err)
			}
	
			mtn.tab[ii] = TrieEntry{
				Type: TRIE_TYPE_MORE,
				Hash: next_hash,
			}
			continue
		}

		if ent0.Type == TRIE_TYPE_MORE && ent1.Type == TRIE_TYPE_ITEM {
			ment, err := mtn.mergeInsert(ent0, ent1)
			if err != nil {
				return mtn, trace(err)
			}

			mtn.tab[ii] = ment
			continue
		}

		if ent1.Type == TRIE_TYPE_MORE && ent0.Type == TRIE_TYPE_ITEM {
			ment, err := mtn.mergeInsert(ent1, ent0)
			if err != nil {
				return mtn, trace(err)
			}

			mtn.tab[ii] = ment
			continue
		}

		if ent0.Type == TRIE_TYPE_ITEM && ent1.Type == TRIE_TYPE_ITEM {
			ment, err := mtn.mergeItems(ent0, ent1)
			if err != nil {
				return mtn, trace(err)
			}

			mtn.tab[ii] = ment
			continue
		}

		panic(fmt.Sprintf("Unhandled case (%d, %d)", ent0.Type, ent1.Type))
	}

	return mtn, nil
}

func (ptn *TrieNode) mergeInsert(ent0, ent1 TrieEntry) (TrieEntry, error) {
	if ent0.Type != TRIE_TYPE_MORE {
		return TrieEntry{}, fmt.Errorf("First argument must be TRIE_TYPE_MORE")
	}

	if ent1.Type != TRIE_TYPE_ITEM {
		return TrieEntry{}, fmt.Errorf("Second argument must be TRIE_TYPE_ITEM")
	}

	mtn := &TrieNode{
		eft: ptn.eft,
		dep: ptn.dep,
	}
		
	var err error

	if !HashesEqual(ent0.Hash, ZERO_HASH) {
		mtn, err = ptn.loadChild(ent0.Hash)
		if err != nil {
			return TrieEntry{}, trace(err)
		}
	}

	key, err := ent1.KeyBytes()
	if err != nil {
		return TrieEntry{}, trace(err)
	}

	err = mtn.insert(key, ent1)
	if err != nil {
		return TrieEntry{}, trace(err)
	}

	hash, err := mtn.save()
	if err != nil {
		return TrieEntry{}, trace(err)
	}

	ment := TrieEntry{
		Type: TRIE_TYPE_MORE,
		Hash: hash,
	}

	return ment, nil
}

func (mtn *TrieNode) mergeItems(ent0, ent1 TrieEntry) (TrieEntry, error) {
	if ent0.Type != TRIE_TYPE_ITEM || ent1.Type != TRIE_TYPE_ITEM {
		return TrieEntry{}, fmt.Errorf("Both arguments must be TRIE_TYPE_ITEM")
	}

	key0, err := ent0.KeyBytes()
	if err != nil {
		return TrieEntry{}, trace(err)
	}
			
	key1, err := ent1.KeyBytes()
	if err != nil {
		return TrieEntry{}, trace(err)
	}
			
	if bytes.Equal(key0, key1) {
		info0, err := mtn.eft.loadItemInfo(ent0.Hash)
		if err != nil {
			return TrieEntry{}, trace(err)
		}

		info1, err := mtn.eft.loadItemInfo(ent1.Hash)
		if err != nil {
			return TrieEntry{}, trace(err)
		}
		
		fmt.Println("XX - Merging", info0.Path, info1.Path)

		if info0.ModT > info1.ModT {
			return ent0, nil
		} else {
			return ent1, nil
		}
	}

	ment := TrieEntry{
		Type: TRIE_TYPE_MORE,
	}

	ment, err = mtn.mergeInsert(ment, ent0)
	if err != nil {
		return TrieEntry{}, trace(err)
	}

	ment, err = mtn.mergeInsert(ment, ent1)
	if err != nil {
		return TrieEntry{}, trace(err)
	}

	return ment, nil
}
