package eft

import (
	"os"
	"time"
	"path"
)

type Checkpoint struct {
	Hash string
	Adds string
	Dels string
}

func (eft *EFT) MakeCheckpoint() (*Checkpoint, error) {
	eft.Lock()
	defer eft.Unlock()

	eft.begin()

	snap := eft.mainSnap()
	when := uint64(time.Now().UnixNano())
	root := HashToHex(snap.Root)

	err := eft.logUpdate(snap, when, "CPT", root)
	if err != nil {
		eft.abort()
		return nil, trace(err)
	}

	dels, err := eft.collect()
	if err != nil {
		eft.abort()
		return nil, trace(err)
	}
	
	eft.commit()

	hash, err := eft.loadSnapsHash()
	if err != nil {
		return nil, trace(err)
	}

	adds := eft.TempName()

	err = os.Rename(path.Join(eft.Dir, "added"), adds)
	if err != nil {
		return nil, trace(err)
	}

	cp := &Checkpoint{
		Hash: HashToHex(hash),
		Adds: adds,
		Dels: dels,
	}

	return cp, nil
}

func (cp *Checkpoint) Cleanup() {
	os.Remove(cp.Adds)
	os.Remove(cp.Dels)
}

