package main

import (
	"fmt"
	"github.com/RyanJarv/little-stitch/lib"
	"io"
)

func main() {
	conn := lib.NewConnection([4]byte{34, 125, 181, 202})

	_, err := io.WriteString(conn, "Hello!")
	if err != nil {
		fmt.Printf("failed to write to connection: %s\n", err)
	}
}
