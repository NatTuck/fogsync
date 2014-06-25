
TOOLS := cat-eft ls-eft get-eft put-eft

all:
	echo $(TOOLS) | xargs -n 1 bash -c '(cd $$0 && go build && cp $$0 ../bin)'

clean:
	echo $(TOOLS) | xargs -n 1 bash -c '(cd $$0 && go clean)'
	rm -f bin/*

test:
	(cd eft && go test)

.PHONY: all clean test
