.PHONY: clean debug dev test


clean:
	@rm -rf ./dracli

dev: clean dracli

debug:
	dlv debug --headless --api-version=2 -l 127.0.0.1:2456 -- query pwState

dracli:
	@go build -o dracli 

test: 
	@go test -v -cover