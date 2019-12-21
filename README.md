# Golang Api Template

A simple project containing the foundations needed for developing an API with golang. Currently all the data is stored within the application memory but can easily be extended to utilise several storage mechanisms.
This template is built upon an application that can manage Friends by exposing endpoints that allow CRUD operations.

## Installation
The application run as a TCP server exposed by a user defined port. It is possible to start the application either through a docker container or as a standalone local binary on your host machine.

### Docker
Before proceeding ensure docker is installed on your system. Also ensure there is a `.env` file in the root directory with the very least `PORT` defined (refer to [example](https://github.com/sayze/golang-template/blob/master/.env.example) file)

Run `docker-compose up --build`

Initial install will take several minutes as it needs to pull in the required docker images.

## Local
The most recent golang suite must be installed on your system.

1. Run `make debug|release` depending on your preference
2. Execute the binary ensuring  `PORT` is specified as command line argument  i.e `PORT ./server`


## TODO
- Add wire DI tool
- Refactor package structure to a more domain driven approach
- Add Sqlite support
