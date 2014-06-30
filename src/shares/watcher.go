package shares

// Watches a directory tree for changes.
// Triggers a callback when such a change occurs.

import (
	"github.com/howeyc/fsnotify"
	"os"
	"path"
	"fmt"
	"../fs"
)

type Watcher struct {
	share    *Share
	fswatch  *fsnotify.Watcher
	changes  chan string
	shutdown chan bool
}

func (ss *Share) startWatcher() *Watcher {
	fswatch, err := fsnotify.NewWatcher()
	fs.CheckError(err)

	ww := &Watcher{
		share:   ss,
		fswatch: fswatch,
		changes: make(chan string, 64),
		shutdown: make(chan bool),
	}

	go ww.watcherLoop()

	return ww
}

func (ww *Watcher) scanTree(scan_path string) {
	ww.share.gotLocalChange(scan_path)

	info, err := os.Lstat(scan_path)
	if err != nil || !info.Mode().IsDir() {
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

func (ww *Watcher) Changed(change string) {
	ww.changes<- change
}

func (ww *Watcher) Shutdown() {
	ww.shutdown<- true
}

func (ww *Watcher) watcherLoop() {
	for {
		select {
		case evt := <-ww.fswatch.Event:
			ww.scanTree(evt.Name)
		case err := <-ww.fswatch.Error:
			fmt.Println("XX - error:", err)
			fs.PanicHere("Giving up")
		case chg := <-ww.changes:
			ww.scanTree(chg)
		case _    = <-ww.shutdown:
			fmt.Println("XX - Shutting down watcher")
			ww.fswatch.Close()
			break
		}
	}
}

