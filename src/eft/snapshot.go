package eft

import (
	"encoding/binary"
	"encoding/hex"
	"bytes"
	"path"
	"fmt"
	"strings"
	"io/ioutil"
	"os"
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
	Uuid [32]byte
	Time uint64
	Name string
}

var NoSnapsFile = fmt.Errorf("Snaps File Not Found")
var ZERO_UUID = [32]byte{}

func (eft *EFT) GetSnap(name string) (*Snapshot, error) {
	snaps, err := eft.loadSnaps()
	if err != nil {
		fmt.Println("Note: Loading snaps failed")
	}

	if len(snaps) == 0 {
		snaps = eft.defaultSnapsList()
	}

	if name == "" {
		return &snaps[0], nil
	}

	for ii := 0; ii < len(snaps); ii++ {
		if snaps[ii].Name == name {
			return &snaps[ii], nil
		}
	}

	return nil, ErrNotFound
}


func (eft *EFT) defaultSnapsList() []Snapshot {
	snaps := []Snapshot{}
	snaps = append(snaps, Snapshot{eft: eft, Uuid: RandomBytes32()})
	return snaps
}

func (eft *EFT) loadSnapsHash() ([32]byte, error) {
	hash := [32]byte{}

	snaps_path := path.Join(eft.Dir, "snaps")
	hash_text, err := ioutil.ReadFile(snaps_path)
	if err != nil {
		return hash, NoSnapsFile
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
	if err == NoSnapsFile {
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
		copy(snap.Uuid[:], data[base + 32:base + 64])
	
		snap.Name = string(data[base + 64:base + 96])
		snap.Name = strings.Trim(snap.Name, "\x00")

		be := binary.BigEndian
		snap.Time = be.Uint64(data[base + 96:base + 104])

		if HashesEqual(ZERO_UUID, snap.Uuid) {
			return nil, fmt.Errorf("Found Zero UUID in snap %d", ii)
		}

		if !bytes.Equal(snap.Root[:], zero_hash) {
			snaps = append(snaps, snap)
		}
	}

	if len(snaps) == 0 {
		snaps = eft.defaultSnapsList()
	}

	return snaps, nil
}

func (snap *Snapshot) Save() error {
	rd := snap.rootsDir()

	return fmt.Errorf("Not Implemented: %s", rd)
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

		if HashesEqual(ZERO_UUID, snap.Uuid) {
			return fmt.Errorf("Found Zero UUID in snap %d", ii)
		}

		copy(data[base:base + 32], snap.Root[:])
		copy(data[base + 32:base + 64], snap.Uuid[:])

		if len(snap.Name) > 31 {
			snap.Name = snap.Name[0:31]
		}

		// Zero pad the end of Desc
		desc_bytes := make([]byte, 32)
		copy(desc_bytes, []byte(snap.Name))

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

func (snap *Snapshot) rootsDir() string {
	root_dir := path.Join(snap.eft.Dir, "tmp", "roots", HashToHex(snap.Uuid))
	os.MkdirAll(root_dir, 0750)
	return root_dir
}

func (snap *Snapshot) debugDump(trie *EFT) {
	fmt.Printf("[Snapshot] %s \n\t@ %s (\"%s\")\n",
	    hex.EncodeToString(snap.Root[:]),
		dateFromUnix(snap.Time), 
		snap.Name)

	pt, err := trie.loadPathTrie(snap.Root)
	if err != nil {
		panic(err)
	}

	pt.debugDump()
}

func (snap *Snapshot) mergeRoots() error {
	//rs, err := snap.listRoots()

	return fmt.Errorf("Not implemented")
}

func (snap *Snapshot) listRoots() ([][32]byte, error) {
	rd := snap.rootsDir()

	infos, err := ioutil.ReadDir(rd)
	if err != nil {
		return nil, trace(err)
	}

	roots := make([][32]byte, 0)

	for _, si := range(infos) {
		name := si.Name()
		if len(name) != 64 {
			continue
		}

		roots = append(roots, HexToHash(name))
	}
	
	return roots, nil
}

