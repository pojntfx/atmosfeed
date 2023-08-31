# Public variables
DESTDIR ?=
PREFIX ?= /usr/local
OUTPUT_DIR ?= out
DST ?=

# Private variables
obj = atmosfeed
classifiers = alwaystrue
all: build

# Build
build: $(addprefix build/,$(obj)) $(addprefix build-classifier/,$(classifiers))
$(addprefix build/,$(obj)):
ifdef DST
	go build -o $(DST) ./cmd/$(subst build/,,$@)
else
	go build -o $(OUTPUT_DIR)/$(subst build/,,$@) ./cmd/$(subst build/,,$@)
endif

build-classifier: $(addprefix build-classifier/,$(classifiers))
$(addprefix build-classifier/,$(classifiers)):
	cd classifiers/$(subst build-classifier/,,$@) && scale function build # && scale signature generate $(subst build-classifier/,,$@):latest
	scale function export local/$(subst build-classifier/,,$@):latest $(OUTPUT_DIR)

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
	rm -rf out pkg/models

# Dependencies
depend:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

	go generate ./...