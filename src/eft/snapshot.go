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
	Root [32]byte
	Log  [32]byte
	Time uint64
	Temp bool
	Desc string
}

func (eft *EFT) defaultSnapsList() ([]Snapshot, error) {
	snaps := []Snapshot{}
	snaps = append(snaps, Snapshot{})
	return snaps, nil
}

func (eft *EFT) loadSnapsHash() ([32]byte, error) {
	hash := [32]byte{}

	snaps_path := path.Join(eft.Dir, "snaps")
	hash_text, err := ioutil.ReadFile(snaps_path)
	if err != nil {
		return hash, ErrNotFound
	}

	hash_slice, err := hex.DecodeString(string(hash_text))
	if err != nil {
		return hash, trace(err)
	}

	copy(hash[:], hash_slice) 
	return hash, nil
}

func (eft *EFT) saveSnapsHash(hash [32]byte) error {
	snaps_path := path.Join(eft.Dir, "snaps")
	hash_text := hex.EncodeToString(hash[:])

	err := ioutil.WriteFile(snaps_path, []byte(hash_text), 0600)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (snap *Snapshot) isEmpty() bool {
	zero := make([]byte, 32)
	return bytes.Compare(zero, snap.Root[:]) == 0
}

func (eft *EFT) loadSnaps() ([]Snapshot, error) {
	hash, err := eft.loadSnapsHash()
	if err == ErrNotFound {
		snaps, err := eft.defaultSnapsList()
		if err != nil {
			return snaps, trace(err)
		}

		return snaps, nil 

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

	for ii := 0; ii < 128; ii++ {
		snap := Snapshot{}
		base := ii * SNAP_SIZE

		copy(snap.Root[:], data[base:base + 32])
		copy(snap.Log[:], data[base + 32:base + 64])
	
		snap.Desc = string(data[base + 64:base + 96])
		snap.Desc = strings.Trim(snap.Desc, "\x00")

		be := binary.BigEndian
		snap.Time = be.Uint64(data[base + 96:base + 104])

		if !bytes.Equal(snap.Root[:], zero_hash) {
			snaps = append(snaps, snap)
		}
	}

	if len(snaps) == 0 {
		snaps, err = eft.defaultSnapsList()
		if err != nil {
			return snaps, trace(err)
		}
	}

	return snaps, nil
}

func (eft *EFT) saveSnaps(snaps []Snapshot) error {
	if len(snaps) == 0 {
		return fmt.Errorf("No snapshots to save")
	}
	
	data := make([]byte, BLOCK_SIZE)

	for ii, snap := range(snaps) {
		if snap.Temp {
			continue
		}

		base := ii * SNAP_SIZE

		copy(data[base:base + 32], snap.Root[:])
		copy(data[base + 32:base + 64], snap.Log[:])

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

func (eft *EFT) mainSnap() *Snapshot {
	if len(eft.Snaps) == 0 {
		snap := Snapshot{}
		eft.Snaps = append(eft.Snaps, snap)
	}

	return &eft.Snaps[0]
}

