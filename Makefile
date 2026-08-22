build:
	go build -ldflags "-X main.version=$(shell git describe --tags --always --dirty)" -o llr

test:
	go test
