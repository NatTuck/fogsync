package cache

import (
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
	writeDirBlocks(blocks, root_bptr)
}

func writeDirBlocks(blocks pio.File, bptr Bptr) {
	dir := loadDirectory(bptr)
	

	for _, ent := range dir {
		switch ent.Type {
		case "dir":

		default:
			writeEntBlocks(blocks, ent)
	}
}

func writeEntBlocks(blocks pio.File, ent DirEnt) {
	
}
