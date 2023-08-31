# Public variables
DESTDIR ?=
PREFIX ?= /usr/local
OUTPUT_DIR ?= out
DST ?=

# Private variables
obj = atmosfeed
signatures = classifier
classifiers = everything
all: build

# Build
build: $(addprefix build/,$(obj)) $(addprefix build/classifier/,$(classifiers))
$(addprefix build/,$(obj)):
ifdef DST
	go build -o $(DST) ./cmd/$(subst build/,,$@)
else
	go build -o $(OUTPUT_DIR)/$(subst build/,,$@) ./cmd/$(subst build/,,$@)
endif

build/classifier: $(addprefix build/classifier/,$(classifiers))
$(addprefix build/classifier/,$(classifiers)):
	scale function build -d ./classifiers/$(subst build/classifier/,,$@)
	scale function export local/$(subst build/classifier/,,$@):latest $(OUTPUT_DIR)

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
	rm -rf out pkg/models pkg/signatures/*/guest pkg/signatures/*/host

# Dependencies
depend: depend/signature depend/sql

depend/signature: $(addprefix depend/signature/,$(signatures))
$(addprefix depend/signature/,$(signatures)):
	scale signature generate $(subst depend/signature/,,$@):latest -d ./pkg/signatures/$(subst depend/signature/,,$@)
	mkdir -p pkg/signatures/$(subst depend/signature/,,$@)/{guest,host} out
	scale signature export local/$(subst depend/signature/,,$@):latest go guest pkg/signatures/$(subst depend/signature/,,$@)/guest
	scale signature export local/$(subst depend/signature/,,$@):latest go host pkg/signatures/$(subst depend/signature/,,$@)/host

depend/sql:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

	go generate ./...