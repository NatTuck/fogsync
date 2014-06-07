
ROOT := $(shell readlink -f ../..)
#TRACE := strace -f
TRACE := /usr/bin/time

test:
	FOGSYNC_ROOT=$(ROOT) $(TRACE) go test
