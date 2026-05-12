validate: static-analysis/lint static-analysis/vulncheck test/unit examples/samplemod/run
	@echo "Validation successful"

examples/samplemod/build:
	mkdir -p build/examples
	go build -o build/examples/samplemod ./examples/samplemod

examples/samplemod/run: examples/samplemod/build
	build/examples/samplemod \
		--duration 20s \
		cli \
		-sample.important 12 \
		-sample.op.test.rate 60 \
		-sample.op.broken.disable \
		-sample.op.unstable.disable

test/unit:
	mkdir -p build
	go test ./... -v -failfast -coverprofile=build/coverage.out

test/unit-json:
	mkdir -p build
	go test ./... -failfast -coverprofile=build/coverage.out -json > build/unit-test-output.json

static-analysis/lint:
	golangci-lint run --fix

static-analysis/vulncheck:
	go tool -modfile tools/go.mod govulncheck ./... || true

static-analysis/vulncheck-sarif:
	mkdir -p build
	go tool -modfile tools/go.mod govulncheck -format sarif ./... > build/vulncheck.sarif

install/tools:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.12.2
