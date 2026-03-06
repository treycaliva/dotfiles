BINARY    := dotfiles-installer
CMD       := ./cmd/installer
GOFLAGS   := -trimpath -ldflags="-s -w"

.PHONY: build install clean test

build:
	go build $(GOFLAGS) -o $(BINARY) $(CMD)

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)

test:
	go test ./...
