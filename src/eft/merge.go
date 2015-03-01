package eft

import (
	"fmt"
	"bytes"
)

func (eft *EFT) MergeRemote(hash [32]byte) error {
	err := eft.with_read_lock(func() {
		currHash, err := eft.getRoot()
		assert_no_error(err)

		if HashesEqual(currHash, hash) {
			fmt.Println("XX - Remote snap has no changes.")
			return
		}

		ptR, err := eft.loadPathTrie(hash)
		assert_no_error(err)
	
		ptL, err := eft.loadPathTrie(currHash)
		assert_no_error(err)

		trie, err := eft.mergePathTries(ptL, ptR)
		assert_no_error(err)

		merged, err := trie.save()
		assert_no_error(err)

		eft.saveRoot(merged)
	})

	if err != nil {
		return err
	}

	return eft.Commit()
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
		tri: tn0.tri,
		dep: tn0.dep,
	}

	for ii := 0; ii < 256; ii++ {
		ent0 := tn0.tab[ii]
		ent1 := tn1.tab[ii]

		if ent0.Hash == ent1.Hash {
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

			next_hash := ent1.Hash

			if !smtn.Equals(stn1) {
				next_hash, err = smtn.save()
				if err != nil {
					return mtn, trace(err)
				}
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
	
	var err error

	mtn := &TrieNode{
		eft: ptn.eft,
		tri: ptn.tri,
		dep: ptn.dep + 1,
	}

	if !HashesEqual(ent0.Hash, ZERO_HASH) {
		mtn, err = ptn.loadChild(ent0.Hash)
		if err != nil {
			return TrieEntry{}, trace(err)
		}
	}

	key := ptn.KeyBytes(ent1)

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

	key0 := mtn.KeyBytes(ent0)
	key1 := mtn.KeyBytes(ent1)
			
	if bytes.Equal(key0, key1) {
		info0 := mtn.eft.loadItemInfo(ent0.Hash)
		info1 := mtn.eft.loadItemInfo(ent1.Hash)
		
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

	ment, err := mtn.mergeInsert(ment, ent0)
	if err != nil {
		return TrieEntry{}, trace(err)
	}

	ment, err = mtn.mergeInsert(ment, ent1)
	if err != nil {
		return TrieEntry{}, trace(err)
	}

	return ment, nil
}

