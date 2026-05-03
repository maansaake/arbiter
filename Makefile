validate: lint unit-test build-examples-samplemod
	@echo "Validation successful"

build-examples-samplemod:
	mkdir -p build/examples
	go build -o build/examples/samplemod ./examples/samplemod

lint:
	golangci-lint run --fix

unit-test:
	go test ./... -failfast

function-test:
	build/examples/samplemod -duration 20s cli -sample.important 12 -sample.op.test.rate 60

vulncheck:
	cd tools/govulncheck && go tool govulncheck -C ../.. ./...
