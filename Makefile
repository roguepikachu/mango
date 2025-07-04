.PHONY: build test e2e run

build:
	go build -o mango ./cmd/mango

test:
	go test ./...

e2e:
        go test -tags=e2e ./...

run: build
	./mango run
