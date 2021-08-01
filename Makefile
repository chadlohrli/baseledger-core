.PHONY: build clean install integration lint migrate mod test

default: build

build: clean mod
	go fmt ./...
	go build -v -o ./.bin/node ./cmd/node
	go build -v -o ./.bin/migrate ./cmd/migrate

clean:
	rm -rf ./.bin 2>/dev/null || true
	go fix ./...
	go clean -i ./...

install: clean
	go install ./...

lint:
	./ops/lint.sh

migrate: mod
	rm -rf ./.bin/migrate 2>/dev/null || true
	go build -v -o ./.bin/migrate ./cmd/migrate
	./ops/migrate.sh

mod:
	go mod init 2>/dev/null || true
	go mod tidy
	go mod vendor

test: build
	./ops/run_local_tests.sh

integration: build
	./ops/run_integration_tests.sh
