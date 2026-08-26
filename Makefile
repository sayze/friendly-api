cd=$(PWD)
path?=...

test: unit

# To run tests across the entire service, a package, or a single test.
# make unit <- run across entire service
# make unit path=internal/friend/service <- run all tests in that package
# make unit path=internal/friend/service test=TestService_Create <- run a single test
unit:
	@if [ -n "$(test)" ]; then \
		echo "Running unit test $(path)/$(test)"; \
		go test $(cd)/$(path) -run $(test) -json --tags unit | tparse; \
	else \
		echo "Running unit tests $(cd)/$(path)"; \
		go test $(cd)/$(path) -json --tags unit | tparse; \
	fi

build:
	@echo "Building cmd binaries..."
	go build -o bin/ ./cmd/api

clean-mocks:
	@echo "Removing existing mocks..."
	find . -type f -name "mock_gen.go" -delete

mocks: clean-mocks
	@echo "Generating project wide mocks..."
	go generate ./...

clean-build:
	@echo "Removing cmd binaries..."
	rm -rf bin

fmt:
	gofumpt -w .
