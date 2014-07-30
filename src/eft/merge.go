package eft

import (
	"io/ioutil"
	"os"
	"fmt"
)

func (eft *EFT) MergeRemote(hash [32]byte, fn func(bs string) error) error {
	// Fetch snaps root
	bs := eft.TempName()
	
	err := ioutil.WriteFile(bs, []byte(fmt.Sprintf("%s\n", HashToHex(hash))), 0600)
	if err != nil {
		return trace(err)
	}
	defer os.Remove(bs)
	
	err = fn(bs)
	if err != nil {
		return trace(err)
	}

	// Merge snapshots
	snaps, err := eft.loadSnaps()
	if err != nil {
		return trace(err)
	}

	rem_snaps, err := eft.loadSnapsFrom(hash)
	if err != nil {
		return trace(err)
	}

	if len(rem_snaps) != 1 {
		panic("TODO: Handle multiple snapshots")
	}

	merged, err = eft.mergeSnaps(loc_snaps[0], rem_snaps[0])
	if err != nil {
		return trace(err)
	}
	snaps[0] = merged

	err = eft.saveSnaps(snaps)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) mergeSnaps(snap0 Snapshot, snap1 Snapshot) (Snapshot, error) {
	if HashesEqual(snap0.Root, snap1.Root) {
		return snap0, nil
	}

	root, err := eft.MergePathTries(snap0.Root, snap1.Root)
	if err != nil {
		return Snapshot{}, err
	}

	snap0.Root = root
	return snap0, nil
}

func (pt *PathTrie) merge(tn0, tn1 *TrieNode) (*TrieNode, error) {
	mtn := &TrieNode{
		eft: tn0.eft,
		key: tn0.key,
	}

	for ii := 0; ii < 256; ii++ {
		ent0 := tn0.tab[ii]
		ent1 := tn1.tab[ii]

		if ent0 == ent1 {
			mtn.tab[ii] = ent0
			continue
		}

		if ent0.Type == TRIE_TYPE_NONE {
			mtn.tab[ii] = ent1
			continue
		}

		if ent1.Type == TRIE_TYPE_NONE {
			mtn.tab[ii] = ent0
		}

		mtn.tab[ii] = ent0.merge(ent1)
	}

	return mtn, nil
}


