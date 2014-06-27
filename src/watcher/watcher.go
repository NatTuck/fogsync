package watcher

// Watches a directory tree for changes.
// Triggers a callback when such a change occurs.

import (
	"github.com/howeyc/fsnotify"
)

OnChangeFn func(changed_path string) error

type Watcher struct {
	Base string
	OnChange 
}

func New(base string, on_change ) *Watcher {
	
}
