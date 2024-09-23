BINARY_NAME=tf-provider-lldap
DIST_DIR=dist

$(info $(shell mkdir -p $(DIST_DIR)))

default: all

all: build test run

build:
	go build -o "${DIST_DIR}/${BINARY_NAME}" main.go

docs: deps
	~/go/bin/tfplugindocs validate
	~/go/bin/tfplugindocs generate

test: build
	./scripts/test.sh

debug:
	DEBUG_LOCAL=yes go run main.go

clean:
	go clean
	go mod tidy
	rm -f "${DIST_DIR}/${BINARY_NAME}"

deps:
	go mod download
	@cat tools/tools.go | grep _ | awk -F '"' '{print $$2}' | xargs -tI {} go install {}
