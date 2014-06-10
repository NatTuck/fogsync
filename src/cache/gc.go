package cache

import (
	"fmt"
	"os"
	"../fs"
	"../config"
	"../pio"
)

func StartGC(share *config.Share) {
	st := StartST(share)
	defer st.Finish()

	if (st.share.Root == "") {
		return
	}

	blocks_name := config.TempName()
	blocks := pio.Create(blocks_name)
	defer func() {
		blocks.Close()
		os.Remove(blocks_name)
	}()

	root_bptr := BptrFromString(st.share.Root)
	st.findLiveBlocks(&blocks, root_bptr)
	
	info, err := os.Lstat(blocks_name)
	fs.CheckError(err)

	blocks.Close()
	fmt.Println("Blocks:", info.Size() / BPTR_SIZE)

	live := OpenBlockList(blocks_name)
	defer live.Close()

	live.Sort()

	fs.FindFiles(share.CacheDir(), func(file_path string) {
		hash, ok := share.PathToHash(file_path)
		if !ok {
			return
		}

		if !live.HasBlock(hash) {
			fmt.Println("Found garbage:", file_path)
			err := os.Remove(file_path)
			fs.CheckError(err)
		}
	})
}

func (st *ST) findLiveBlocks(blocks *pio.File, bptr Bptr) {
	dir := st.loadDirectory(bptr)

	blocks.Write(bptr.Hash)

	for _, ent := range dir {
		bptr1 := BptrFromString(ent.Bptr)

		if ent.Type == "dir" {
			st.findLiveBlocks(blocks, bptr1)
		}

		temp_name := st.decryptFromBlocks1(bptr1, blocks)
		os.Remove(temp_name)
	}
}
