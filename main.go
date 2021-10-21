package main

import (
	"fmt"
	"github.com/RyanJarv/little-stitch/lib"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	if (len(os.Args) >= 1) && (os.Args[1] == "client") {
		var ip [4]byte
		if len(os.Args) < 3 {
			fmt.Printf("USAGE: %s client <ip>\n", os.Args[0])
			os.Exit(2)
		}
		split := strings.Split(os.Args[2], ".")
		fmt.Println(split)
		if len(split) != 4 {
			fmt.Printf("USAGE: %s client <ip>\n", os.Args[0])
			os.Exit(2)
		}
		for i := 0; i < 4; i++ {
			d, err := strconv.Atoi(split[i])
			if err != nil {
				fmt.Printf("failed to parse ip address: %s\n", err)
			}
			ip[i] = byte(d)
		}
		conn := lib.NewConnection(ip)
		_, err := io.WriteString(conn, "Hello!")
		if err != nil {
			fmt.Printf("failed to write to connection: %s\n", err)
			os.Exit(3)
		}
	} else if (len(os.Args) >= 1) && (os.Args[1] == "server") {
		r, err := lib.NewReceiver()
		if err != nil {
			fmt.Printf("failed to write to connection: %s\n", err)
			os.Exit(3)
		}
		_, err = io.Copy(os.Stdout, r.Out)
		if err != nil {
			log.Fatalln("error copying output to stdout")
		}
	} else {
		fmt.Printf("USAGE: %s <server|client>\n", os.Args[0])
		os.Exit(1)
	}
}
