FROM golang:latest

WORKDIR /usr/src/little-stitch
ENV GO11MODULES on
COPY ./ ./
RUN go build -o /usr/local/bin/little-stitch ./main.go
ENTRYPOINT ["/usr/local/bin/little-stitch"]
