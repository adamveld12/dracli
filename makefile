.PHONY: check clean debug dev test

ci: clean setup lint test

clean:
	@rm -rf ./dracli

dev: clean dracli

debug:
	dlv debug --headless --api-version=2 -l 127.0.0.1:2456 -- query pwState

dracli:
	@go build -o dracli

lint:
	go get golang.org/x/lint/golint
	golint -min_confidence 1.5 -set_exit_status
	go vet -all -v

setup:
	@go get -t -d -v ./...

test:
	@go test -v -cover