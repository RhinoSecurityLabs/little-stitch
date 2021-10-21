FROM golang:latest

WORKDIR /usr/src/little-stitch
ENV GO11MODULES on
RUN apt-get update && apt-get install -y \
    libpcap-dev \
    iproute2
COPY ./go.mod ./go.sum ./*.go ./
RUN go build -o /usr/local/bin/little-stitch ./main.go
ENTRYPOINT ["/usr/local/bin/little-stitch"]
