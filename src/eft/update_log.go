package eft

import (
	"path"
	"fmt"
	"os"
)

func (eft *EFT) logUpdate(snap *Snapshot, etime uint64, etype string, edata string) (eret error) {
	log_name := path.Join(eft.Dir, "updates")

	info, err := eft.loadItem(snap.Log, log_name)
	if err != nil {
		// First log entry.
	}

	flags := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	log, err := os.OpenFile(log_name, flags, 0600)
	if err != nil {
		return trace(err)
	}
	defer func() {
		err := log.Close()
		if eret == nil && err != nil {
			eret = trace(err)
		}
	}()

	line := fmt.Sprintf("%d\t%s\t%s\n", etime, etype, edata)
	_, err = log.WriteString(line)
	if err != nil {
		return trace(err)
	}

	info, err = FastItemInfo(log_name)
	if err != nil {
		return trace(err)
	}

	hash, err := eft.saveItem(info, log_name)
	if err != nil {
		return trace(err)
	}

	snap.Log = hash

	return nil
}


