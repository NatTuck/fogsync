package eft

import (
	"path"
	"io/ioutil"
	"os"
	"sync"
)

// An EFT has one or more "snapshots", each representing a trie root
// that shouldn't be garbage collected.

// Standard snapshot names:
//  - "main" is the local state of the EFT
//  - "remote" is the last successful upload, this is
//    to avoid block downloads during merge   
//  - "upload" is the current upload attempt, this is to
//    avoid a race condition on upload

func (eft *EFT) getSnapRoot(name string) ([32]byte, error) {
	root_path := path.Join(eft.Dir, "snaps", "main")

	root_hex, err := ReadOneLine(root_path)
	if err != nil {
		return ZERO_HASH, trace(err)
	}

	return HexToHash(root_hex), nil
}

func (eft *EFT) getRoot() ([32]byte, error) {
	return eft.getSnapRoot("main")
}

func (eft *EFT) putSnapRoot(name string, hash [32]byte) {
	if eft.locked != LOCKED_RW {
		panic(ErrNeedLock)
	}

	err := os.MkdirAll(path.Join(eft.Dir, "snaps"), 0750)
	assert_no_error(err)

	snap_root := path.Join(eft.Dir, "snaps", name)

	err = WriteOneLine(snap_root, HashToHex(hash))
	assert_no_error(err)
}

func (eft *EFT) rootsDir() string {
	root_dir := path.Join(eft.Dir, "tmp", "roots")
	
	err := os.MkdirAll(root_dir, 0750)
	assert_no_error(err)

	return root_dir
}

func (eft *EFT) cleanupRoot(old_root, new_root [32]byte) {
	deads, _ := eft.root_changes(old_root, new_root)
	eft.removeBlocks(deads)

	eft.removeRoot(old_root)
}

func (eft *EFT) removeRoot(root [32]byte) {
	root_file := path.Join(eft.rootsDir(), HashToHex(root))

	err := os.Remove(root_file)
	assert_no_error(err)
}

func (eft *EFT) saveRoot(root [32]byte) {
	root_file := path.Join(eft.rootsDir(), HashToHex(root))

	err := WriteOneLine(root_file, HashToHex(root))
	assert_no_error(err)
}

func (eft *EFT) mergeRootPair(r0 [32]byte, r1 [32]byte) [32]byte {
	pt0, err := eft.loadPathTrie(r0)
	assert_no_error(err)

	pt1, err := eft.loadPathTrie(r1)
	assert_no_error(err)

	ptM, err := eft.mergePathTries(pt0, pt1)
	assert_no_error(err)

	root, err := ptM.save()
	assert_no_error(err)

	eft.saveRoot(root)

	return root
}

func (eft *EFT) mergeRoots() [32]byte {
	assert(eft.locked == LOCKED_RW, "Need RW Lock")

	var roots [][32]byte

	for {
		roots = eft.listRoots()

		if len(roots) < 2 {
			break
		}
		
		jobs := len(roots) / 2

		wg := sync.WaitGroup{}
		var eret error = nil

		for ii := 0 ; ii < jobs; ii++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() {
					err := recover_assert()
					if err != nil {
						eret = err
					}
				}()

				r0 := roots[2*ii]
				r1 := roots[2*ii + 1]

				rN := eft.mergeRootPair(r0, r1)

				eft.cleanupRoot(r0, rN)
				eft.cleanupRoot(r1, rN)
			}()

			wg.Wait()
			assert_no_error(eret)
		}
	}

	if len(roots) == 1 {
		prev_root, err := eft.getRoot()
		if err == nil {
			eft.putSnapRoot("main", roots[0])
			eft.removeRoot(roots[0])
			return
		}
			
		root := eft.mergeRootPair(prev_root, roots[0])
		eft.putSnapRoot("main", root)

		eft.removeRoot(roots[0])
		eft.removeRoot(root)
	}
}

func (eft *EFT) listRoots() [][32]byte {
	rd := eft.rootsDir()

	infos, err := ioutil.ReadDir(rd)
	assert_no_error(err)

	roots := make([][32]byte, 0)

	for _, si := range(infos) {
		name := si.Name()
		if len(name) != 64 {
			continue
		}

		roots = append(roots, HexToHash(name))
	}
	
	return roots
}

func (eft *EFT) liveBlocks(name string) *BlockSet {
	root, err := eft.getSnapRoot(name)
	assert_no_error(err)

	trie, err := eft.loadPathTrie(root)
	assert_no_error(err)

	bs := trie.blockSet()
	bs.Add(root)

	return bs
}

func (eft *EFT) listSnaps() []string {
	snaps_dir := path.Join(eft.Dir, "snaps")
	
	infos, err := ioutil.ReadDir(snaps_dir)
	assert_no_error(err)

	snaps := make([]string, 0)

	for _, si := range(infos) {
		name := si.Name()
		
		if len(name) < 3 {
			continue
		}

		snaps = append(snaps, name)
	}

	return snaps
}
