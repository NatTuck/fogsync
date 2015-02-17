package eft

import (
	"fmt"
)

type Checkpoint struct {
	Trie *EFT
	Adds string
	Dels string
}

// Checkpoint does the following:
//  - Merges new roots for each snapshot.
//  - Determines difference between prev and next states.
//    - Which blocks have been added and need uploading?
//    - Which blocks are now garbage and need deleting?
//    - How many blocks are in the new state?
//  - Collects the garbage.

func (eft *EFT) Checkpoint(prev_hash [32]byte) ([32]byte, *Checkpoint, error) {
	eft.Lock()
	defer eft.Unlock()

	dels, err := eft.collect()
	if err != nil {
		return ZERO_HASH, nil, trace(err)
	}

	hash, err := eft.loadSnapsHash()
	if err != nil {
		return ZERO_HASH, nil, trace(err)
	}

	cp := &Checkpoint{
		Trie: eft,
		Adds: "",
		Dels: "",
	}

	fmt.Println("dels:", dels)
	fmt.Println("hash:", hash)

	return ZERO_HASH, cp, nil
}


