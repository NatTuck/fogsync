package eft

import (
	"encoding/binary"
	"encoding/hex"
	"bytes"
	"path"
	"fmt"
	"strings"
	"io/ioutil"
)

// 128 snapshots can be stored in one block
//
// Adding more than 16 snapshots is disallowed to allow EFTs 
// with additional snapshots to be merged in.
//
// If merging would result in more than 128 snapshots, some
// remote snapshots will be discarded.

const MAX_SNAPS = 16
const SNAP_SIZE = 128

type Snapshot struct {
	eft  *EFT
	Root [32]byte
	Time uint64
	Desc string
}

func (eft *EFT) defaultSnapsList() []Snapshot {
	snaps := []Snapshot{}
	snaps = append(snaps, Snapshot{eft: eft})
	return snaps
}

func (eft *EFT) loadSnapsHash() ([32]byte, error) {
	hash := [32]byte{}

	snaps_path := path.Join(eft.Dir, "snaps")
	hash_text, err := ioutil.ReadFile(snaps_path)
	if err != nil {
		return hash, ErrNotFound
	}

	hash_string := strings.Trim(string(hash_text), "\n")

	hash_slice, err := hex.DecodeString(hash_string)
	if err != nil {
		return hash, trace(err)
	}

	copy(hash[:], hash_slice) 
	return hash, nil
}

func (eft *EFT) saveSnapsHash(hash [32]byte) error {
	snaps_path := path.Join(eft.Dir, "snaps")
	hash_text := hex.EncodeToString(hash[:])

	err := ioutil.WriteFile(snaps_path, []byte(hash_text + "\n"), 0600)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (snap *Snapshot) isEmpty() bool {
	zero := make([]byte, 32)
	return bytes.Compare(zero, snap.Root[:]) == 0
}

func (snap *Snapshot) pathTrie() (PathTrie, error) {
	return PathTrie{}, nil
}

func (eft *EFT) loadSnaps() ([]Snapshot, error) {
	hash, err := eft.loadSnapsHash()
	if err == ErrNotFound {
		return eft.defaultSnapsList(), nil 
	} else if err != nil {
		return nil, trace(err)
	}

	snaps, err := eft.loadSnapsFrom(hash)
	if err != nil {
		return nil, trace(err)
	}
	
	return snaps, nil
}

func (eft *EFT) loadSnapsFrom(hash [32]byte) ([]Snapshot, error) {
	data, err := eft.loadBlock(hash)
	if err != nil {
		return nil, trace(err)
	}

	snaps := make([]Snapshot, 0)
	zero_hash := make([]byte, 32)

	for ii := 0; ii < 64; ii++ {
		snap := Snapshot{eft: eft}
		base := ii * SNAP_SIZE

		copy(snap.Root[:], data[base:base + 32])
	
		snap.Desc = string(data[base + 64:base + 96])
		snap.Desc = strings.Trim(snap.Desc, "\x00")

		be := binary.BigEndian
		snap.Time = be.Uint64(data[base + 96:base + 104])

		if !bytes.Equal(snap.Root[:], zero_hash) {
			snaps = append(snaps, snap)
		}
	}

	if len(snaps) == 0 {
		snaps = eft.defaultSnapsList()
	}

	return snaps, nil
}

func (eft *EFT) saveSnaps(snaps []Snapshot) error {
	if len(snaps) == 0 {
		return fmt.Errorf("No snapshots to save")
	}

	prev_snaps, err := eft.loadSnaps()
	if err != nil {
		return trace(err)
	}

	if len(snaps) == len(prev_snaps) {
		snaps_changed := false

		for ii := 0; ii < len(snaps); ii++ {
			if snaps[ii] != prev_snaps[ii] {
				snaps_changed = true
			}
		}

		if !snaps_changed {
			return nil
		}
	}

	data := make([]byte, DATA_SIZE)

	for ii, snap := range(snaps) {
		base := ii * SNAP_SIZE

		copy(data[base:base + 32], snap.Root[:])

		if len(snap.Desc) > 31 {
			snap.Desc = snap.Desc[0:31]
		}

		// Zero pad the end of Desc
		desc_bytes := make([]byte, 32)
		copy(desc_bytes, []byte(snap.Desc))

		copy(data[base + 64:base + 96], desc_bytes)

		be := binary.BigEndian
		be.PutUint64(data[base + 96:base + 104], snap.Time)
	}


	hash, err := eft.saveBlock(data)
	if err != nil {
		return trace(err)
	}

	err = eft.saveSnapsHash(hash)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) getSnap(nn uint) *Snapshot {
	snaps, err := eft.loadSnaps()
	if err != nil {
		fmt.Println("Note: Loading snaps failed")
	}

	if len(eft.Snaps) == 0 {
		snaps = eft.defaultSnapsList()
	}

	return &snaps[nn]
}

func (snap *Snapshot) debugDump(trie *EFT) {
	fmt.Printf("[Snapshot] %s \n\t@ %s (\"%s\")\n",
	    hex.EncodeToString(snap.Root[:]),
		dateFromUnix(snap.Time), 
		snap.Desc)

	pt, err := trie.loadPathTrie(snap.Root)
	if err != nil {
		panic(err)
	}

	pt.debugDump()
}
