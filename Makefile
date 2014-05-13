make:
	go build

fmt:
	go fmt
	golint .

test:
	go test -v
