package eft

import (
)

func (eft *EFT) PrepUpload() {

}

func (eft *EFT) CommitUpload() {

}

func (eft *EFT) root_changes(r0 [32]byte, r1 [32]byte) (*BlockSet, *BlockSet) {
	// Given two path trie roots, find the blocks unique to each.

	pt0, err := eft.loadPathTrie(r0)
	assert_no_error(err)

	pt1, err := eft.loadPathTrie(r1)
	assert_no_error(err)

	bs0 := pt0.blockSet()
	bs1 := pt1.blockSet()

	un0 := bs0.Diff(bs1) // Blocks unique to bs0
	un1 := bs1.Diff(bs0) // Blocks unique to bs1

	return un0, un1
}
