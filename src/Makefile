
VERSION=$(shell echo `date "+%Y-%m-%d %H:%M"` `whoami`@`hostname`)

all: bin/fogsync assets

bin/fogsync: config/version.go $(wildcard */*.go)
	(cd fogsync && go build)
	mkdir -p bin
	cp fogsync/fogsync bin

config/version.go:
	echo "package config" > $@
	echo "var VERSION = \"$(VERSION)\"" >> $@

assets:
	(cd webui/assets && make)

clean:
	(cd fogsync && go clean)
	rm -f bin/fogsync
	(cd webui/assets && make clean)

prereqs:
	go get github.com/howeyc/fsnotify
	go get github.com/GeertJohan/go.rice
	go get github.com/edsrzf/mmap-go
	go get code.google.com/p/go.crypto/nacl/secretbox
	(cd webui/assets && make prereqs)

local-reset:
	rm -rf ~/.cache/fogsync
	rm -rf ~/.config/fogsync

live-test:
	rm -rf ~/.cache/fogsync/Documents
	make
	bin/fogsync

.PHONY: all clean assets config/version.go
