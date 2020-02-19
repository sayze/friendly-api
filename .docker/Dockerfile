FROM golang:latest

WORKDIR /app

RUN go get github.com/githubnemo/CompileDaemon

EXPOSE 4040

ENTRYPOINT CompileDaemon -command="bin/debug/server" -build="make debug" -exclude-dir=bin