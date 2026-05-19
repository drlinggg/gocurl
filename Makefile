BIN    := gocurl
GOBIN  := $(shell go env GOPATH)/bin

build: test
	go build -o $(BIN) .

install: test
	go install .
	@if ! grep -qF '$(GOBIN)' ~/.zshrc 2>/dev/null; then \
		printf '\nexport PATH="$$PATH:$(GOBIN)"\n' >> ~/.zshrc; \
		echo "→ added $(GOBIN) to ~/.zshrc"; \
		echo "→ run: source ~/.zshrc"; \
	else \
		echo "→ $(GOBIN) already in ~/.zshrc"; \
	fi
	@echo "→ gocurl installed to $(GOBIN)"

test:
	go test ./...

clean:
	rm -f $(BIN)
