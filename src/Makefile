

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

live-test:
	rm -rf ~/.cache/fogsync/Documents
	make
	bin/fogsync

.PHONY: all clean assets
