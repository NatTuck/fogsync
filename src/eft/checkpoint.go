package eft

import (
	"os"
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

