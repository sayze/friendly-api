# Common build variables.
BASEPATH = bin
OUTPUT = server
DEBUG = $(BASEPATH)/debug/$(OUTPUT)
RELEASE = $(BASEPATH)/release/$(OUTPUT)
SRC = ./

# Default.
all: debug release

# Release.
release:
	GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-s -w" -o $(RELEASE) $(SRC)

# Debug.
debug:
	go build -o $(DEBUG) $(SRC)

# Run unit test.  
test:
	go test ./internal/...

# Start container in interactive mode.
start:
	docker-compose up

# Stop container.
stop:
	docker-compose stop

# Build container.
build:
	docker-compose build

clean:
	rm -rf bin/