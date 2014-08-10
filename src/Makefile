

all: bin/fogsync assets

bin/fogsync: $(wildcard */*.go)
	(cd fogsync && go build)
	mkdir -p bin
	cp fogsync/fogsync bin

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

live-test:
	rm -rf ~/.cache/fogsync/Documents
	make
	bin/fogsync

.PHONY: all clean assets
