BINARY_NAME=terraform-provider-lldap
DIST_DIR=dist

$(info $(shell mkdir -p $(DIST_DIR)))

default: all

all: lint modupdate build test docs clean

build:
	go build -o "${DIST_DIR}/${BINARY_NAME}" main.go

lint:
	tfproviderlint -R001=false ./...

modupdate:
	go get -u
	go mod tidy

docs:
	tfplugindocs generate

# Run a specific test like this: make test TEST=test-directory-name
test: build
	./scripts/test.sh $(TEST)

debug:
	DEBUG_LOCAL=yes go run main.go

clean:
	go clean
	go mod tidy
	rm -f "${DIST_DIR}/${BINARY_NAME}"

.PHONY: all build lint modupdate docs test debug clean
