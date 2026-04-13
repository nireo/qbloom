.PHONY: test fix vet lint

test:
	go test ./...

fix:
	gofmt -w *.go

vet:
	go vet ./...

lint: fix vet test
