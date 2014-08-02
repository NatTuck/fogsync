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

func (ss *Share) startWatcher() *Watcher {
	fswatch, err := fsnotify.NewWatcher()
	fs.CheckError(err)

	ww := &Watcher{
		share:   ss,
		fswatch: fswatch,
		updates:  make(chan string, 64),
		shutdown: make(chan bool),
	}

	go ww.watcherLoop()

	return ww
}

func (ww *Watcher) scanTree(scan_path string) {
	sysi, err := os.Lstat(scan_path)
	if err == nil {
		ww.share.gotLocalUpdate(scan_path, sysi)
	}
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

func (ww *Watcher) checkEft(update_path string) {
	ww.share.gotRemoteUpdate(update_path)
}

func (ww *Watcher) Changed(change string) {
	ww.updates<- change
}

func (ww *Watcher) ChangedRemote(change string) {
	ww.remotes<- change
}

func (ww *Watcher) Shutdown() {
	ww.shutdown<- true
}

func (ww *Watcher) watcherLoop() {
	for {
		select {
		case evt := <-ww.fswatch.Event:
			if evt.IsDelete() || evt.IsRename() {
				stamp := uint64(time.Now().UnixNano())
				ww.share.gotLocalDelete(evt.Name, stamp)
			} else {
				ww.scanTree(evt.Name)
			}
		case err := <-ww.fswatch.Error:
			fmt.Println("XX - error:", err)
			fs.PanicHere("Giving up")
		case upd := <-ww.updates:
			ww.scanTree(upd)
		case upd := <-ww.remotes:
			ww.checkEft(upd)
		case _    = <-ww.shutdown:
			fmt.Println("XX - Shutting down watcher")
			ww.fswatch.Close()
			break
		}
	}
}

