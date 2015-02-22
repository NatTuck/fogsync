package eft

import (
	"bytes"
	"errors"
	"fmt"
)

type ByteTrie interface {
	KeyBytes(ee TrieEntry) ([]byte, error)
}

const (
	TRIE_TYPE_NONE = 0
	TRIE_TYPE_MORE = 1
	TRIE_TYPE_OVRF = 2
	TRIE_TYPE_ITEM = 3
)

type TrieEntry struct {
	Hash [32]byte
	Pkey [8]byte
	Data [6]byte
	Type uint8
	Resv uint8
}

type TrieNode struct {
	eft *EFT
	tri ByteTrie
	dep int

	hdr [2048]byte
	ovr [16][32]byte
	tab [256]TrieEntry 
}

var ErrNotFound = errors.New("EFT: record not found")

func (tn *TrieNode) KeyBytes(ee TrieEntry) ([]byte, error) {
	if tn.tri == nil {
		return nil, fmt.Errorf("Invalid TrieNode (no tri)")
	}
	return tn.tri.KeyBytes(ee)
}

func (tn *TrieNode) Equals(tn1 *TrieNode) bool {
	for ii := 0; ii < 16; ii++ {
		if tn.ovr[ii] != tn1.ovr[ii] {
			return false
		}
	}

	for ii := 0; ii < 256; ii++ {
		if tn.tab[ii] != tn1.tab[ii] {
			return false
		}
	}

	return true
}

func (tn *TrieNode) emptyChild() *TrieNode {
	return &TrieNode{
		eft: tn.eft,
		tri: tn.tri,
		dep: tn.dep + 1,
	}
}

func (tn *TrieNode) loadChild(hash [32]byte) (*TrieNode, error) {
	cc := tn.emptyChild()

	err := cc.load(hash)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

func (tn *TrieNode) load(hash [32]byte) error {
	data, err := tn.eft.loadBlock(hash)
	if err != nil {
		return trace(err)
	}

	copy(tn.hdr[:], data[0:2048])

	for ii := 0; ii < 256; ii++ {
		offset := 2048 + 48 * ii
		rec := data[offset:offset + 48]

		entry := TrieEntry{}
		copy(entry.Hash[:], rec[0:32])
		entry.Type = rec[32]
		copy(entry.Pkey[:], rec[34:42])
		copy(entry.Data[:], rec[42:48])
		tn.tab[ii] = entry
	}

	for ii := 0; ii < 16; ii++ {
		offset := 14336 + 32 * ii
		copy(tn.ovr[ii][:], data[offset:offset + 32])
	}

	return nil
}

func (tn *TrieNode) save() ([32]byte, error) {
	data := make([]byte, DATA_SIZE)

	copy(data[0:2048], tn.hdr[:])

	for ii := 0; ii < 256; ii++ {
		offset := 2048 + 48 * ii
		rec := data[offset:offset + 48]

		entry := tn.tab[ii]
		copy(rec[0:32], entry.Hash[:])
		rec[32] = entry.Type
		copy(rec[34:42], entry.Pkey[:])
		copy(rec[42:48], entry.Data[:])
	}

	for ii := 0; ii < 16; ii++ {
		offset := 14336 + 32 * ii
		copy(data[offset:offset + 32], tn.ovr[ii][:])
	}

	hash, err := tn.eft.saveBlock(data)
	if err != nil {
		return hash, trace(err)
	}

	return hash, nil
}

func (tn *TrieNode) find(key []byte) ([32]byte, error) {
	slot := key[tn.dep]
	entry := tn.tab[slot]

	switch entry.Type {
	case TRIE_TYPE_NONE:
		return [32]byte{}, ErrNotFound

	case TRIE_TYPE_MORE:
		next, err := tn.loadChild(entry.Hash)
		if err != nil {
			return [32]byte{}, err // Could be ErrNotFound, no trace
		}

		return next.find(key)

	case TRIE_TYPE_ITEM:
		key1, err := tn.tri.KeyBytes(entry)
		if err != nil {
			return [32]byte{}, trace(err)
		}

		if bytes.Compare(key, key1) == 0 {
			return entry.Hash, nil
		} else {
			return [32]byte{}, ErrNotFound
		}

	default:
		return [32]byte{}, trace(fmt.Errorf("Unknown type in node entry: %d", entry.Type))
	}
}

func (tn *TrieNode) insert(key []byte, new_ent TrieEntry) error {
	slot := key[tn.dep]
	entry := tn.tab[slot]

	new_ent.Type = TRIE_TYPE_ITEM

	switch entry.Type {
	case TRIE_TYPE_NONE:
		// Insert into empty slot.
		tn.tab[slot] = new_ent

	case TRIE_TYPE_ITEM:
		curr_key, err := tn.tri.KeyBytes(entry)
		if err != nil {
			return trace(err)
		}
		
		if bytes.Compare(key, curr_key) == 0 {
			// Replace
			tn.tab[slot] = new_ent

		} else {
			// Push down

			next := tn.emptyChild()

			err = next.insert(curr_key, entry)
			if err != nil {
				return trace(err)
			}
	
			err = next.insert(key, new_ent)
			if err != nil {
				return trace(err)
			}
	
			next_hash, err := next.save()
			if err != nil {
				return trace(err)
			}

			next_entry := TrieEntry{ Type: TRIE_TYPE_MORE }
			next_entry.Hash = next_hash

			tn.tab[slot] = next_entry
		}

	case TRIE_TYPE_MORE:
		next, err := tn.loadChild(entry.Hash)
		if err != nil {
			return trace(err)
		}

		err = next.insert(key, new_ent)
		if err != nil {
			return trace(err)
		}

		next_hash, err := next.save()
		if err != nil {
			return trace(err)
		}

		entry.Hash = next_hash
		tn.tab[slot] = entry

	default:
		return trace(fmt.Errorf("Invalid entry type: %d", entry.Type))
	}

	return nil
}

func (tn *TrieNode) remove(key []byte) error {
	slot := key[tn.dep]
	entry := tn.tab[slot]

	switch entry.Type {
	case TRIE_TYPE_NONE:
		return ErrNotFound

	case TRIE_TYPE_ITEM:
		tn.tab[slot] = TrieEntry{}

		fmt.Println("TODO (EFT): Figure out merge on remove")

	case TRIE_TYPE_MORE:
		next, err := tn.loadChild(entry.Hash)
		if err != nil {
			return trace(err)
		}

		err = next.remove(key)
		if err != nil {
			return err
		}

		hash, err := next.save()
		if err != nil {
			return trace(err)
		}

		entry.Hash = hash
		tn.tab[slot] = entry

	default:
		return trace(fmt.Errorf("Invalid entry type: %d", entry.Type))
	}

	return nil
}

func (tn *TrieNode) visitEachEntry(fn func(ent *TrieEntry) error) error {
	for ii := 0; ii < 256; ii++ {
		ent := &tn.tab[ii]

		if ent.Type == TRIE_TYPE_NONE {
			continue
		}

		err := fn(ent)
		if err != nil {
			return trace(err)
		}

		if ent.Type == TRIE_TYPE_MORE {
			next, err := tn.loadChild(ent.Hash)
			if err != nil {
				return err
			}

			err = next.visitEachEntry(fn)
			if err != nil {
				return trace(err)
			}
		}

		if ent.Type == TRIE_TYPE_OVRF {
			panic("TODO: Handle overflow tables")
		}
	}

	return nil
}
