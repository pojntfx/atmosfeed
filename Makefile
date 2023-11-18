# Public variables
DESTDIR ?=
PREFIX ?= /usr/local
OUTPUT_DIR ?= out
DST ?=

WWWROOT ?= /var/www/html
WWWPREFIX ?= /atmosfeed

# Private variables
clis = atmosfeed-server atmosfeed-client
pwas = frontend
signatures = $(notdir $(wildcard pkg/signatures/*))
all: build

# Build
build: build/cli build/pwa

build/cli: $(addprefix build/cli/,$(clis))
$(addprefix build/cli/,$(clis)):
ifdef DST
	go build -o $(DST) ./cmd/$(subst build/cli/,,$@)
else
	go build -o $(OUTPUT_DIR)/$(subst build/cli/,,$@) ./cmd/$(subst build/cli/,,$@)
endif

build/pwa:
	cd frontend && bun run build
	mkdir -p $(OUTPUT_DIR)
	tar -cvzf $(OUTPUT_DIR)/frontend.tar.gz -C frontend/out .

# Install
install: install/cli install/pwa

install/cli: $(addprefix install/cli/,$(clis))
$(addprefix install/cli/,$(clis)):
	install -D -m 0755 $(OUTPUT_DIR)/$(subst install/cli/,,$@) $(DESTDIR)$(PREFIX)/bin/$(subst install/cli/,,$@)

install/pwa:
	mkdir -p $(DESTDIR)$(WWWROOT)$(WWWPREFIX)
	tar -xvf out/frontend.tar.gz -C $(DESTDIR)$(WWWROOT)$(WWWPREFIX)

# Uninstall
uninstall: uninstall/cli uninstall/pwa

uninstall/cli: $(addprefix uninstall/cli/,$(clis))
$(addprefix uninstall/cli/,$(clis)):
	rm $(DESTDIR)$(PREFIX)/bin/$(subst uninstall/cli/,,$@)

uninstall/pwa:
	rm -rf $(DESTDIR)$(WWWROOT)$(WWWPREFIX)

# Run
run: run/cli run/pwa

run/cli: $(addprefix run/cli/,$(clis))
$(addprefix run/cli/,$(clis)):
	$(subst run/cli/,,$@) $(ARGS)

run/pwa:
	cd frontend && bun run start

# Test
test: test/cli test/pwa

test/cli:
	go test -timeout 3600s -parallel $(shell nproc) ./...

test/pwa:
	cd frontend && bun run test

# Benchmark
benchmark: benchmark/cli benchmark/pwa

benchmark/cli:
	go test -timeout 3600s -bench=./... ./...

benchmark/pwa:
	cd frontend && bun run test

# Clean
clean: clean/cli clean/pwa

clean/cli:
	rm -rf out pkg/models pkg/signatures/*/go/*

clean/pwa:
	rm -rf frontend/node_modules frontend/.next frontend/out

# Dependencies
depend: depend/cli depend/pwa depend/signature

depend/cli:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

	go generate ./...

depend/pwa:
	cd frontend && bun install

depend/signature: $(addprefix depend/signature/,$(signatures))
$(addprefix depend/signature/,$(signatures)):
	scale --no-telemetry signature generate $(subst depend/signature/,,$@):latest -d ./pkg/signatures/$(subst depend/signature/,,$@)
	mkdir -p ./pkg/signatures/$(subst depend/signature/,,$@)/go/host
	scale --no-telemetry signature export local/$(subst depend/signature/,,$@):latest go host ./pkg/signatures/$(subst depend/signature/,,$@)/go/host
