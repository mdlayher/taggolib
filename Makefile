.PHONY: fmt test bench

make:
	go build

fmt:
	go fmt
	golint .

test:
	go test -v

bench:
	go test -v -run=NONE -bench=.
