
ROOT := $(shell readlink -f ../..)
#TRACE := strace -f
TRACE := /usr/bin/time

app:
	FOGSYNC_ROOT=$(ROOT) go build

test:
	FOGSYNC_ROOT=$(ROOT) $(TRACE) go test

.PHONY: app test
