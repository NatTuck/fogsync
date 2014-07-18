package eft

func (eft *EFT) loadItemInfo(hash []byte) (ItemInfo, error) {
	info := ItemInfo{}

	data, err := eft.loadBlock(hash)
	if err != nil {
		return info, err
	}

	info = ItemInfoFromBytes(data[0:2048])

	return info, nil
}

func (eft *EFT) loadItem(hash []byte, dst_path string) (ItemInfo, error) {
	info, err := eft.loadItemInfo(hash)
	if err != nil {
		return info, err
	}

	if info.Size <= 12 * 1024 {
		return eft.loadSmallItem(hash, dst_path)
	} else {
		return eft.loadLargeItem(hash, dst_path)
	}
}

func (eft *EFT) saveItem(info ItemInfo, src_path string) ([]byte, error) {
	if (info.Size <= 12 * 1024) {
		return eft.saveSmallItem(info, src_path)
	} else {
		return eft.saveLargeItem(info, src_path)
	}
}

func (eft *EFT) visitItemBlocks(info ItemInfo, fn func(hash []byte) error) error {
	if (info.Size <= 12 * 1024) {
		return nil
	} else {
		trie, err := eft.loadLargeTrie(info.Hash[:])
		if err != nil {
			return trace(err)
		}

		return trie.visitEachBlock(fn)
	}
}

