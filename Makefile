validate: lint vulncheck unit-test build-examples-samplemod function-test
	@echo "Validation successful"

build-examples-samplemod:
	mkdir -p build/examples
	go build -o build/examples/samplemod ./examples/samplemod

lint:
	golangci-lint run --fix

unit-test:
	go test ./... -failfast

vulncheck:
	go tool -modfile tools/go.mod govulncheck ./...

function-test: build-examples-samplemod
	build/examples/samplemod -duration 20s cli -sample.important 12 -sample.op.test.rate 60 -sample.op.broken.disable -sample.op.unstable.disable