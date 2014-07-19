package eft

import (
	"time"
	"fmt"
	"os"
)

func (eft *EFT) logUpdate(snap *Snapshot, etime time.Time, etype string, edata string) (eret error) {
	log_name := eft.TempName()

	info, err := eft.loadItem(snap.Log, log_name)
	if err != nil {
		// First log entry.
	}
	defer os.Remove(log_name)

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

	stamp := uint64(etime.UnixNano())

	line := fmt.Sprintf("%d\t%s\t%s\n", stamp, etype, edata)
	_, err = log.WriteString(line)
	if err != nil {
		return trace(err)
	}

	info.ModT = stamp

	hash, err := eft.saveItem(info, log_name)
	if err != nil {
		return trace(err)
	}

	snap.Log = hash

	return nil
}


