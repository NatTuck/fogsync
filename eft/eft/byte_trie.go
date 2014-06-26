package eft

import (
	"errors"
	"fmt"
)

const (
	TRIE_TYPE_NONE = 0
	TRIE_TYPE_MORE = 1
	TRIE_TYPE_ITEM = 2
)

type TrieEntry struct {
	Hash [32]byte
	Data [14]byte
	Type uint8
	Byte uint8
}

type NodeShop interface {
	getKey(ee TrieEntry) ([]byte, error)
}

type TrieNode struct {
	eft *EFT
	shp NodeShop
	par *TrieNode

	hdr [4096]byte
	tab [256]TrieEntry 
}

var ErrNotFound = errors.New("EFT: record not found")

func (tn *TrieNode) emptyChild() *TrieNode {
	return &TrieNode{
		eft: tn.eft,
		shp: tn.shp,
		par: tn,
	}
}

func (tn *TrieNode) loadChild(hash []byte) (*TrieNode, error) {
	cc := tn.emptyChild()

	err := cc.load(hash)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

func (tn *TrieNode) load(hash []byte) error {
	data, err := tn.eft.loadBlock(hash)
	if err != nil {
		return trace(err)
	}

	copy(tn.hdr[:], data[0:4096])

	base := 4 * 1024

	for ii := 0; ii < 256; ii++ {
		offset := base + 48 * ii
		rec := data[offset:offset + 48]

		entry := TrieEntry{}
		copy(entry.Hash[:], rec[0:32])
		entry.Type = rec[32]
		entry.Byte = rec[33]
		copy(entry.Data[:], rec[34:48])
		tn.tab[ii] = entry
	}

	return nil
}

func (tn *TrieNode) save() ([]byte, error) {
	data := make([]byte, BLOCK_SIZE)

	copy(data[0:4096], tn.hdr[:])

	base := 4 * 1024

	for ii := 0; ii < 256; ii++ {
		offset := base + 48 * ii
		rec := data[offset:offset + 48]

		entry := tn.tab[ii]
		copy(rec[0:32], entry.Hash[:])
		rec[32] = entry.Type
		rec[33] = entry.Byte
		copy(rec[34:48], entry.Data[:])
	}

	hash, err := tn.eft.saveBlock(data)
	if err != nil {
		return nil, trace(err)
	}

	return hash, nil
}

func (tn *TrieNode) find(key []byte, dd int) ([]byte, error) {
	slot := key[dd]
	entry := tn.tab[slot]

	switch entry.Type {
	case TRIE_TYPE_NONE:
		return nil, ErrNotFound
	case TRIE_TYPE_MORE:
		next_hash := entry.Hash[:]

		next, err := tn.loadChild(next_hash)
		if err != nil {
			return nil, err // Could be ErrNotFound, no trace
		}

		return next.find(key, dd + 1)
	case TRIE_TYPE_ITEM:
		key1, err := tn.shp.getKey(entry)
		if err != nil {
			return nil, trace(err)
		}

		if BytesEqual(key, key1) {
			return entry.Hash[:], nil
		} else {
			return nil, ErrNotFound
		}
	default:
		return nil, trace(fmt.Errorf("Unknown type in node entry: %d", entry.Type))
	}
}

func (tn *TrieNode) insert(key []byte, newEnt TrieEntry, dd int) error {
	slot := key[dd]
	entry := tn.tab[slot]

	switch entry.Type {
	case TRIE_TYPE_NONE:
		// Insert into empty slot.
		tn.tab[slot] = newEnt

	case TRIE_TYPE_ITEM:
		curr_key, err := tn.shp.getKey(entry)
		if err != nil {
			return trace(err)
		}
		
		if dd > 0 {
			newEnt.Byte = key[dd - 1]
		}

		if BytesEqual(key, curr_key) {
			// Replace
			err := tn.eft.killItemBlocks(entry.Hash[:])
			if err != nil {
				return trace(err)
			}
			
			tn.tab[slot] = newEnt
		} else {
			return tn.insertConflict(key, newEnt, dd)
		}

	case TRIE_TYPE_MORE:
		next, err := tn.loadChild(entry.Hash[:])
		if err != nil {
			return trace(err)
		}

		err = next.insert(key, newEnt, dd + 1)
		if err != nil {
			return trace(err)
		}

		next_hash, err := next.save()
		if err != nil {
			return trace(err)
		}

		// We need to update all matching hashes to
		// handle merged tables.
		for ii := 0; ii < 256; ii++ {
			if BytesEqual(tn.tab[ii].Hash[:], entry.Hash[:]) {
				copy(tn.tab[ii].Hash[:], next_hash)
			}
		}
	default:
		return trace(fmt.Errorf("Invalid entry type: %d", entry.Type))
	}

	return nil
}

func (tn *TrieNode) insertConflict(key []byte, newEnt TrieEntry, dd int) error {
	slot := key[dd]
	entry := tn.tab[slot]

	if tn.par != nil && entry.Byte != newEnt.Byte {
		// This is a merged table conflict. Time to split.
		node := tn.par.emptyChild()

		for ii := 0; ii < 256; ii++ {
			if tn.tab[ii].Byte == newEnt.Byte {
				node.tab[ii] = tn.tab[ii]
				tn.tab[ii] = TrieEntry{}
			}
		}

		hash, err := node.save()
		if err != nil {
			return trace(err)
		}

		copy(tn.par.tab[newEnt.Byte].Hash[:], hash)

		err = tn.par.mergeChild(newEnt.Byte)
		if err != nil {
			return trace(err)
		}

		return nil
	}

	next := tn.emptyChild()

	// Is a merge possible?
	entryKey, err := tn.shp.getKey(entry)
	if err != nil {
		return trace(err)
	}

	if entryKey[dd + 1] != key[dd + 1] {
		// Can we find a merge candidate?
		for ii := 0; ii < 256; ii++ {
			if tn.tab[ii].Type != TRIE_TYPE_MORE {
				continue
			}

			cc, err := tn.loadChild(tn.tab[ii].Hash[:])
			if err != nil {
				return trace(err)
			}
			
			if cc.tab[slot].Type != TRIE_TYPE_NONE {
				continue
			}

			// Awesome, we can merge.
			next = cc
		}
	}

	err = next.insert(key, entry, dd + 1)
	if err != nil {
		return trace(err)
	}
	
	err = next.insert(key, newEnt, dd + 1)
	if err != nil {
		return trace(err)
	}
	
	next_hash, err := next.save()
	if err != nil {
		return trace(err)
	}
			
	entry.Type = TRIE_TYPE_MORE
	copy(entry.Hash[:], next_hash)
	tn.tab[slot] = entry

	return nil
}

func (tn *TrieNode) remove(key []byte, dd int) error {
	slot := key[dd]
	entry := tn.tab[slot]

	switch entry.Type {
	case TRIE_TYPE_NONE:
		return ErrNotFound

	case TRIE_TYPE_ITEM:
		err := tn.eft.killItemBlocks(entry.Hash[:])
		if err != nil {
			return trace(err)
		}

		tn.tab[slot] = TrieEntry{}

		fmt.Println("TODO: Figure out merge on remove")

	case TRIE_TYPE_MORE:
		next, err := tn.loadChild(entry.Hash[:])
		if err != nil {
			return trace(err)
		}

		err = next.remove(key, dd + 1)
		if err != nil {
			return err
		}

	default:
		return trace(fmt.Errorf("Invalid entry type: %d", entry.Type))
	}

	return nil
}

func (tn *TrieNode) bitMap() bitMap {
	bm := bitMap{}

	for ii := 0; ii < 256; ii++ {
		if tn.tab[ii].Type != TRIE_TYPE_NONE {
			bm.set(uint8(ii), true)
		}
	}

	return bm
}

func (tn *TrieNode) mergeChild(slot uint8) error {
	alice, err := tn.loadChild(tn.tab[slot].Hash[:])
	if err != nil {
		return trace(err)
	}

	a_full := alice.bitMap()

	for ii := 0; ii < 256; ii++ {
		if tn.tab[ii].Type != TRIE_TYPE_MORE {
			continue
		}

		bob, err := tn.loadChild(tn.tab[ii].Hash[:])
		if err != nil {
			return trace(err)
		}

		b_full := bob.bitMap()

		if !a_full.canMergeWith(b_full) {
			continue
		}

		// We can merge with bob
		fmt.Println("TODO: Implement merge")
		return nil
	}

	return nil
}

