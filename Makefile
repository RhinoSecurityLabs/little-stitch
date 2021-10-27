BINARY_NAME=little-stitch

build:
	GOARCH=amd64 GOOS=darwin go build -ldflags "-s -w" -o ${BINARY_NAME}-darwin-amd64 main.go
	GOARCH=arm64 GOOS=darwin go build -ldflags "-s -w" -o ${BINARY_NAME}-darwin-arm64 main.go

docker/build:
	docker run -e GO11MODULES=on -w /build -v "${PWD}:/build" golang:latest make build

run:
	./${BINARY_NAME}

build_and_run: build run

clean:
	go clean
	rm ${BINARY_NAME}-darwin-amd64
	rm ${BINARY_NAME}-darwin-arm64

test:
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

dep:
	go mod download

vet:
	go vet

lint:
	golangci-lint run --enable-all
