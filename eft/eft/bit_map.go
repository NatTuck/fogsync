package eft

// This implements a 256-bit bit map to determine occupancy in trie
// nodes.

type bitMap struct {
	bits [4]uint64
}

func (bm *bitMap) set(nn uint8, vv bool) {
	ii := nn / 64
	jj := nn % 64

	if vv {
		bm.bits[ii] |= 1 << jj
	} else {
		bm.bits[ii] &^= 1 << jj
	}
}

func (bm *bitMap) get(nn uint8) bool {
	ii := nn / 64
	jj := nn % 64

	return 0 != (bm.bits[ii] & 1) << jj
}

func (bm *bitMap) canMergeWith(bm1 bitMap) bool {
	for ii := 0; ii < 4; ii++ {
		if 0 != bm.bits[ii] & bm1.bits[ii] {
			return false
		}
	}

	return true
}

