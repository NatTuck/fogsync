package eft

import (
	"os"
	"path"
)

type Checkpoint struct {
	Trie *EFT
	Hash string
	Adds string
	Dels string
}

func (eft *EFT) MakeCheckpoint() (*Checkpoint, error) {
	eft.Lock()

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
		Trie: eft,
		Hash: HashToHex(hash),
		Adds: adds,
		Dels: dels,
	}

	return cp, nil
}

func (cp *Checkpoint) Abort() {
	defer cp.Trie.Unlock()
	os.Remove(cp.Dels)
	os.Rename(cp.Adds, path.Join(cp.Trie.Dir, "added"))
}

func (cp *Checkpoint) Commit() {
	defer cp.Trie.Unlock()
	os.Remove(cp.Adds)
	os.Remove(cp.Dels)
}

