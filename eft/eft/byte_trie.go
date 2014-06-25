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

type TrieNode struct {
	eft *EFT
	hdr [4096]byte
	tab [256]TrieEntry 
}

type GetKeyFn func(TrieEntry) ([]byte, error)

var ErrNotFound = errors.New("EFT: record not found")

func (eft *EFT) loadTrieNode(hash []byte) (*TrieNode, error) {
	data, err := eft.LoadBlock(hash)
	if err != nil {
		return nil, trace(err)
	}
	
	tn := &TrieNode{ eft: eft }
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

	return tn, nil
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

func (tn *TrieNode) find(key []byte, dd int) ([]byte, error) {
	slot := key[dd]
	entry := node.tab[slot]

	switch entry.Type {
	case TRIE_TYPE_NONE:
		return nil, ErrNotFound
	case TRIE_TYPE_MORE:
		next_hash := tn.tab[slot].Hash[:]

		next, err := tn.eft.loadTrieNode(next_hash)
		if err != nil {
			return nil, err // Could be ErrNotFound, no trace
		}

		return next.find(ii, dd + 1)
	case TRIE_TYPE_DATA:
		return ent.Hash[:], nil
	default:
		return nil, trace(fmt.Errorf("Unknown type in node entry: %d", ent.Type))
	}
}

func (tn *TrieNode) insert(key []byte, newEnt TrieEntry, getKey GetKeyFn, dd int) error {
	slot := key[dd]
	entry := tn.tab[slot]

	return nil
}

