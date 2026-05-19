BIN    := gocurl
GOBIN  := $(shell go env GOPATH)/bin
SHELL_RC := $(HOME)/.$(notdir $(SHELL))rc

.PHONY: build install uninstall test run vet fmt clean

build: test
	go build -o $(BIN) .

install: test
	go install .
	@if ! grep -qF '$(GOBIN)' $(SHELL_RC) 2>/dev/null; then \
		printf '\nexport PATH="$$PATH:$(GOBIN)"\n' >> $(SHELL_RC); \
		echo "→ added $(GOBIN) to $(SHELL_RC)"; \
		echo "→ run: source $(SHELL_RC)"; \
	else \
		echo "→ $(GOBIN) already in $(SHELL_RC)"; \
	fi
	@echo "→ gocurl installed to $(GOBIN)"

uninstall:
	rm -f $(GOBIN)/$(BIN)
	@if [ -f $(SHELL_RC) ] && grep -qF 'export PATH="$$PATH:$(GOBIN)"' $(SHELL_RC); then \
		grep -vF 'export PATH="$$PATH:$(GOBIN)"' $(SHELL_RC) > $(SHELL_RC).tmp && \
		mv $(SHELL_RC).tmp $(SHELL_RC); \
		echo "→ removed $(GOBIN) from $(SHELL_RC)"; \
	fi
	@echo "→ gocurl uninstalled from $(GOBIN)"

test:
	go test ./...

run:
	go run . $(ARGS)

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -f $(BIN)
