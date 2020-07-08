# Friendly Api
Backend service implementation for [friendly](https://github.com/sayze/friendly).

## Installation
The application is run as a TCP server exposed by a user defined port. It is possible to start the application either through a docker container or as a standalone local binary on your host machine.

### Docker
Before proceeding ensure docker has been installed on your system.

Run `docker build  .` OR `docker build --build-arg PORT=1234 .` for a custom port.

Initial install will take several minutes as it needs to pull in the required docker images.

## Local
The most recent golang suite must be installed on your system.

1. Run `go build`
2. Execute the resulting binary 
