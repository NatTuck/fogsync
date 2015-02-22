package eft

import (
	"encoding/binary"
	"encoding/hex"
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

var ZERO_UUID = [32]byte{}

func (eft *EFT) GetSnap(name string) (*Snapshot, error) {
	snaps, err := eft.loadSnaps()
	if err != nil {
		return nil, trace(err)
	}

	if name == "" {
		return snaps[0], nil
	}

	for ii := 0; ii < len(snaps); ii++ {
		if snaps[ii].Name == name {
			return snaps[ii], nil
		}
	}

	return nil, ErrNotFound
}


func (eft *EFT) defaultSnaps() []*Snapshot {
	snaps := []*Snapshot{}
	return append(snaps, &Snapshot{eft: eft, Uuid: RandomBytes32()})
}

func (eft *EFT) loadSnapsHash() ([32]byte, error) {
	hash := [32]byte{}

	snaps_path := path.Join(eft.Dir, "snaps")
	hash_text, err := ReadOneLine(snaps_path)
	if err != nil {
		// Pass through ENOENT
		return hash, err
	}

	hash_slice, err := hex.DecodeString(hash_text)
	if err != nil {
		return hash, trace(err)
	}

	copy(hash[:], hash_slice) 
	return hash, nil
}

func (eft *EFT) saveSnapsHash(hash [32]byte) error {
	snaps_path := path.Join(eft.Dir, "snaps")
	hash_text := hex.EncodeToString(hash[:])

	err := WriteOneLine(snaps_path, hash_text)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (snap *Snapshot) isEmpty() bool {
	return HashesEqual(ZERO_HASH, snap.Root)
}

func (snap *Snapshot) pathTrie() (PathTrie, error) {
	return PathTrie{}, nil
}

func (eft *EFT) loadSnaps() ([]*Snapshot, error) {
	if eft.locked == LOCKED_NO {
		return nil, ErrNeedLock
	}

	hash, err := eft.loadSnapsHash()
	if os.IsNotExist(err) {
		// Try again with a write lock to avoid
		// a race on picking the first UUID.
		eft.Unlock()
		eft.Lock()
	
		hash, err = eft.loadSnapsHash()
		if os.IsNotExist(err) {
			fmt.Println("Note: No snaps file. Writing defaults.") 
			snaps := eft.defaultSnaps()
			
			err = eft.saveSnaps(snaps)
			if err != nil {
				return nil, trace(err)
			}
		
			return snaps, nil
		}
	} 
	if err != nil {
		return nil, trace(err)
	}

	snaps, err := eft.loadSnapsFrom(hash)
	if err != nil {
		return nil, trace(err)
	}

	if len(snaps) == 0 {
		snaps = eft.defaultSnaps()
	}
	
	return snaps, nil
}

func (eft *EFT) loadSnapsFrom(hash [32]byte) ([]*Snapshot, error) {
	data, err := eft.loadBlock(hash)
	if err != nil {
		return nil, trace(err)
	}

	snaps := make([]*Snapshot, 0)

	for ii := 0; ii < 64; ii++ {
		snap := &Snapshot{eft: eft}
		base := ii * SNAP_SIZE

		copy(snap.Root[:], data[base:base + 32])
		copy(snap.Uuid[:], data[base + 32:base + 64])
	
		snap.Name = string(data[base + 64:base + 96])
		snap.Name = strings.Trim(snap.Name, "\x00")

		be := binary.BigEndian
		snap.Time = be.Uint64(data[base + 96:base + 104])


		if !HashesEqual(ZERO_UUID, snap.Uuid) {
			snaps = append(snaps, snap)
		}
	}

	return snaps, nil
}

func (snap *Snapshot) saveRoot(root [32]byte) error {
	rd := snap.rootsDir()

	name := HashToHex(root)
	rofn := path.Join(rd, name)

	err := WriteOneLine(rofn, name)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (eft *EFT) saveSnaps(snaps []*Snapshot) error {
	if eft.locked == LOCKED_NO {
		return ErrNeedLock
	}

	if len(snaps) == 0 {
		return fmt.Errorf("No snapshots to save")
	}

	if !HashesEqual(ZERO_HASH, snaps[0].Root) {
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

	pt.debugDump(1)
}

func (snap *Snapshot) removeRoot(root [32]byte) error {
	root_file := path.Join(snap.rootsDir(), HashToHex(root))

	err := os.Remove(root_file)
	if err != nil {
		return trace(err)
	}

	return nil
}

func (snap *Snapshot) mergeRootPair(r0 [32]byte, r1 [32]byte) ([32]byte, error) {
	pt0, err := snap.eft.loadPathTrie(r0)
	if err != nil {
		return ZERO_HASH, trace(err)
	}

	pt1, err := snap.eft.loadPathTrie(r1)
	if err != nil {
		return ZERO_HASH, trace(err)
	}

	ptM, err := snap.eft.mergePathTries(pt0, pt1)
	if err != nil {
		return ZERO_HASH, trace(err)
	}

	root, err := ptM.save()
	if err != nil {
		return ZERO_HASH, trace(err)
	}

	err = snap.saveRoot(root)
	if err != nil {
		return ZERO_HASH, trace(err)
	}

	return root, nil
}

func (snap *Snapshot) mergeRoots() error {
	if snap.eft.locked != LOCKED_RW {
		return ErrNeedLock
	}

	var roots [][32]byte
	var err   error

	for {
		roots, err = snap.listRoots()
		if err != nil {
			return trace(err)
		}
		if len(roots) < 2 {
			break
		}
		
		jobs := len(roots) / 2

		for ii := 0 ; ii < jobs; ii++ {
			r0 := roots[2*ii]
			r1 := roots[2*ii + 1]

			_, err := snap.mergeRootPair(r0, r1)
			if err != nil {
				return trace(err)
			}

			err = snap.removeRoot(r0)
			if err != nil {
				return trace(err)
			}
			
			err = snap.removeRoot(r1)
			if err != nil {
				return trace(err)
			}
		}
	}

	if len(roots) == 1 {
		root, err := snap.mergeRootPair(snap.Root, roots[0])
		if err != nil {
			return trace(err)
		}

		snap.Root = root

		err = snap.eft.commitSnap(snap)
		if err != nil {
			return trace(err)
		}

		err = snap.removeRoot(roots[0])
		if err != nil {
			return trace(err)
		}

		err = snap.removeRoot(root)
		if err != nil {
			return trace(err)
		}
	}

	return nil
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

func (eft *EFT) commitSnap(snap *Snapshot) error {
	if eft.locked != LOCKED_RW {
		return ErrNeedLock
	}

	snaps, err := eft.loadSnaps()
	if err != nil {
		return trace(err)
	}

	var prev_snap *Snapshot = nil

	for _, sn := range(snaps) {
		if HashesEqual(snap.Uuid, sn.Uuid) {
			prev_snap = sn
		}
	}

	if prev_snap == nil {
		return fmt.Errorf("No matching snap found")
	}

	prev_snap.Root = snap.Root

	return eft.saveSnaps(snaps)
}
