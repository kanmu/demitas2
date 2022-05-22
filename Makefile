.PHONY: all
all: vet build

.PHONY: build
build:
	go build ./cmd/dmts

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	rm -rf dmts dmts.exe
