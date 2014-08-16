package shares

// Watches a directory tree for changes.
// Triggers a callback when such a change occurs.

import (
	"github.com/howeyc/fsnotify"
	"os"
	"path"
	"fmt"
	"time"
	"../fs"
)

type Watcher struct {
	share    *Share
	fswatch  *fsnotify.Watcher
	updates  chan string
	remotes  chan string
	shutdown chan bool
}

func (ww *Watcher) Changed(change string) {
	ww.updates<- change
}

func (ww *Watcher) ChangedRemote(change string) {
	fmt.Println("XX - ChangedRemote(", change, ")")
	ww.remotes<- change
}

func (ww *Watcher) Shutdown() {
	ww.shutdown<- true
}

func (ss *Share) startWatcher() *Watcher {
	fswatch, err := fsnotify.NewWatcher()
	fs.CheckError(err)

	ww := &Watcher{
		share:   ss,
		fswatch: fswatch,
		updates:  make(chan string, 64),
		remotes:  make(chan string, 64),
		shutdown: make(chan bool),
	}

	go ww.watcherLoop()

	return ww
}

func (ww *Watcher) scanTree(scan_path string) {
	ww.share.gotChange(scan_path)

	sysi, err := os.Lstat(scan_path)
	if err != nil || !sysi.Mode().IsDir() {
		return
	}

	ww.fswatch.Watch(scan_path)

	dir, err := os.Open(scan_path)
	fs.CheckError(err)

	ents, err := dir.Readdir(-1)
	fs.CheckError(err)

	for _, ent := range(ents) {
		next_path := path.Join(scan_path, ent.Name())
		ww.scanTree(next_path)
	}
}

func (ww *Watcher) watcherLoop() {
	for {
		fmt.Println("XX - Watcher in select")

		select {
		case evt := <-ww.fswatch.Event:
			if evt == nil {
				fmt.Println("XX - Watcher nil event")
				goto DONE
			}

			fmt.Println("XX - Watcher fswatch event")

			if evt.IsDelete() || evt.IsRename() {
				stamp := uint64(time.Now().UnixNano())
				ww.share.gotDelete(evt.Name, stamp)
			} else {
				ww.scanTree(evt.Name)
			}
		case err := <-ww.fswatch.Error:
			if err != nil {
				fmt.Println("XX - error:", err)
			}
			goto DONE
		case upd := <-ww.updates:
			fmt.Println("XX - Watcher Local Update", upd)
			ww.scanTree(upd)
		case upd := <-ww.remotes:
			fmt.Println("XX - Watcher Remote Update", upd)
			full_path := ww.share.FullPath(upd)
			ww.share.gotChange(full_path)
		case _    = <-ww.shutdown:
			fmt.Println("XX - Shutting down watcher")
			goto DONE
		}
	}

  DONE:
	ww.fswatch.Close()
	ww.share.WaitGr.Done()
}

