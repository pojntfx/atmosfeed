# Public variables
DESTDIR ?=
PREFIX ?= /usr/local
OUTPUT_DIR ?= out
DST ?=

# Private variables
obj = atmosfeed
signatures = classifier
functions = alwaystrue
all: build

# Build
build: $(addprefix build/,$(obj)) $(addprefix build/function/,$(functions))
$(addprefix build/,$(obj)):
ifdef DST
	go build -o $(DST) ./cmd/$(subst build/,,$@)
else
	go build -o $(OUTPUT_DIR)/$(subst build/,,$@) ./cmd/$(subst build/,,$@)
endif

build/function: $(addprefix build/function/,$(functions))
$(addprefix build/function/,$(functions)):
	scale function build -d ./pkg/functions/$(subst build/function/,,$@)
	scale function export local/$(subst build/function/,,$@):latest $(OUTPUT_DIR)

# Install
install: $(addprefix install/,$(obj))
$(addprefix install/,$(obj)):
	install -D -m 0755 $(OUTPUT_DIR)/$(subst install/,,$@) $(DESTDIR)$(PREFIX)/bin/$(subst install/,,$@)

# Uninstall
uninstall: $(addprefix uninstall/,$(obj))
$(addprefix uninstall/,$(obj)):
	rm $(DESTDIR)$(PREFIX)/bin/$(subst uninstall/,,$@)

# Run
$(addprefix run/,$(obj)):
	$(subst run/,,$@) $(ARGS)

# Test
test:
	go test -timeout 3600s -parallel $(shell nproc) ./...

# Benchmark
benchmark:
	go test -timeout 3600s -bench=./... ./...

# Clean
clean:
	rm -rf out pkg/models pkg/signatures

depend/signature: $(addprefix depend/signature/,$(signatures))
$(addprefix depend/signature/,$(signatures)):
	scale signature generate $(subst depend/signature/,,$@):latest -d ./api/signatures/$(subst depend/signature/,,$@)
	mkdir -p pkg/signatures/$(subst depend/signature/,,$@)/{guest,host}
	scale signature export local/$(subst depend/signature/,,$@):latest go guest pkg/signatures/$(subst depend/signature/,,$@)/guest
	scale signature export local/$(subst depend/signature/,,$@):latest go host pkg/signatures/$(subst depend/signature/,,$@)/host

# Dependencies
depend: depend/signature
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

	go generate ./...