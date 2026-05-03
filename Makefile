validate: lint unit-test build-examples-samplemod
	@echo "Validation successful"

build-examples-samplemod:
	mkdir -p build/examples
	go build -o build/examples/samplemod ./examples/samplemod

lint:
	golangci-lint run --fix

unit-test:
	go test ./... -failfast
