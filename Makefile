DIST_DIR=dist

$(info $(shell mkdir -p $(DIST_DIR)))

default: all

all: lint modupdate build test docs clean

build:
	go build -o "${DIST_DIR}/terraform-provider-lldap" cmd/terraform-provider-lldap/main.go
	go build -o "${DIST_DIR}/lldap-cli" cmd/lldap-cli/main.go

lint:
	tfproviderlint -R001=false ./...
	golangci-lint run lldap
	golangci-lint run cmd/lldap-cli
	golangci-lint run cmd/terraform-provider-lldap

modupdate:
	go get -u ./...
	go mod tidy

docs:
	tfplugindocs generate --provider-dir cmd/terraform-provider-lldap
	rm -Rf docs
	mv cmd/terraform-provider-lldap/docs .

test: build
	MAKE_TERMOUT=1 ./scripts/test.sh all

unittest: unittest-lldap unittest-cli

unittest-lldap: build
	go test -v ./lldap

unittest-cli: build
	go test -v ./cmd/lldap-cli

inttest: inttest-lldap inttest-cli inttest-terraform

inttest-lldap: build
	MAKE_TERMOUT=1 ./scripts/test.sh inttest-lldap

inttest-cli: build
	MAKE_TERMOUT=1 ./scripts/test.sh inttest-cli

inttest-terraform: build
	MAKE_TERMOUT=1 TEST=$(TEST) ./scripts/test.sh inttest

debug:
	DEBUG_LOCAL=yes go run cmd/terraform-provider-lldap/main.go
	DEBUG_LOCAL=yes go run cmd/lldap-cli/main.go

clean:
	go clean
	go mod tidy
	rm -f "${DIST_DIR}/terraform-provider-lldap" "${DIST_DIR}/lldap-cli"

.PHONY: all build lint modupdate docs test unittest unittest-lldap unittest-cli inttest inttest-lldap inttest-cli inttest-terraform debug clean
