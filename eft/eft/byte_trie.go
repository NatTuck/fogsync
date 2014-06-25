package eft

import (
	"encoding/binary"
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
	Meta [16]byte
}

type GetKeyFn func(TrieEntry) ([]byte, error)

type TrieNode struct {
	eft *EFT

	key GetKeyFn
	dep int

	hdr [4096]byte
	tab [256]TrieEntry 
}

var ErrNotFound = errors.New("EFT: record not found")

func (eft *EFT) loadTrieRoot(hash []byte, getKey GetKeyFn) (*TrieNode, error) {
	tn := &TrieNode{ 
		eft: eft,
		key: getKey,
		dep: 0,
	}

	err := tn.load(hash)
	if err != nil {
		return nil, err
	}

	return tn, nil
}

func (tn *TrieNode) loadNext(hash []byte) (*TrieNode, error) {
	next := &TrieNode{
		eft: tn.eft,
		key: tn.key,
		dep: tn.dep + 1,
	}

	err := next.load(hash)
	if err != nil {
		return nil, err
	}

	return next, nil
}

func (tn *TrieNode) load(hash []byte) error {
	data, err := eft.LoadBlock(hash)
	if err != nil {
		return nil, trace(err)
	}

	copy(tn.hdr[:], data[0:4096])

	base := 4 * 1024
	be := binary.BigEndian

	for ii := 0; ii < 256; ii++ {
		offset := base + 48 * ii
		rec := data[offset:offset + 48]

		ent := TrieEntry{}
		copy(ent.Hash[:], rec[0:32])
		copy(ent.Meta[:], rec[32:48])
		tn.tab[ii] = ent
	}

	return nil
}

func (tn *TrieNode) save() ([]byte, error) {
	data := make([]byte, BLOCK_SIZE)

	copy(data[0:4096], tn.hdr[:])

	base := 4 * 1024
	be := binary.BigEndian

	for ii := 0; ii < 256; ii++ {
		offset := base + 48 * ii
		rec := data[offset:offset + 48]

		ent := tn.tab[ii]
		copy(rec[0:32], ent.Hash[:])
		copy(rec[32:48], ent.Meta[:])
	}

	hash, err := eft.SaveBlock(data)
	if err != nil {
		return nil, trace(err)
	}

	return hash, nil
}

func (tn *TrieNode) find(key []byte) ([]byte, error) {
	slot := key[dd]
	entry := tn.tab[slot]

	switch entry.Type {
	case TRIE_TYPE_NONE:
		return nil, ErrNotFound
	case TRIE_TYPE_MORE:
		next_hash := entry.Hash[:]

		next, err := tn.loadNext(next_hash)
		if err != nil {
			return nil, err // Could be ErrNotFound, no trace
		}

		return next.find(key)
	case TRIE_TYPE_DATA:
		return ent.Hash[:], nil
	default:
		return nil, trace(fmt.Errorf("Unknown type in node entry: %d", ent.Type))
	}
}

func (tn *TrieNode) insert(key []byte, newEnt TrieEntry) error {
	slot := key[dd]
	entry := tn.tab[slot]

	switch entry.Type {
	case TRIE_TYPE_NONE:
		tb.Table[slot] = newEnt
		return nil
	case TRIE_TYPE_DATA:
		if BytesEqual(key, tn.key(entry)) {
			// Replace
			tn.eft.killItemBlocks(entry.Hash[:])

			tb.Table[slot] = newEnt

			return nil
		} else {
			// Push down
			next := &TrieNode{ 
				eft: eft,
				key: getKey,
				dep: tn.dep + 1,
			}

			err := next.insert(key, entry)
			if err != nil {
				return trace(err)
			}

			err = next.insert(key, newEnt)
			if err != nil {
				return trace(err)
			}

			next_hash, err := next.save()
			if err != nil {
				return trace(err)
			}

			tb.Table[slot] = 

			return nil
		}
	case TRIE_TYPE_MORE:
		next_hash := entry.Hash[:]

		next, err := tn.loadNext(next_hash)
		if err != nil {
			return nil, trace(err)
		}

		return next.insert(key, newEnt)
	default:
		return trace(fmt.Errorf("Invalid entry type: %d", entry.Type))
	}
}

func (tn *TrieNode) remove(key []byte) error {
	
}
