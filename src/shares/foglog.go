package shares

import (
	"path"
	"os"
	"fmt"
	"../fs"
)

func (ss *Share) logPath() string {
	return path.Join(ss.CacheDir(), "foglog")
}

func (ss *Share) logEvent(etype string, stamp uint64, rel_path string) {
	flags := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	log, err := os.OpenFile(ss.logPath(), flags, 0600)
	fs.CheckError(err)
	defer func() {
		err := log.Close()
		fs.CheckError(err)
	}()

	log.Write([]byte(fmt.Sprintf("%d\t%s\t%s\n", stamp, etype, rel_path)))
}
