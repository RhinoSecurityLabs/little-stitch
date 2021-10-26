package main

import (
	"flag"
	"fmt"
	"github.com/RyanJarv/little-stitch/lib"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)


func help() {
	log.Fatalf("USAGE: %s <server|client>\n", os.Args[0])
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		help()
	}

	if flag.Args()[0] == "client" {
		client(flag.Args()[1:]...)
	} else if (len(os.Args) >= 1) && (flag.Args()[0] == "server") {
		server()
	} else {
		help()
	}
}

func server() {
	server, err := lib.NewServer()
	if err != nil {
		log.Fatalf("starting server: %s\n", err)
	}

	go func() {
		 _, err = io.Copy(server.Send, os.Stdin)
		 if err != nil {
			 log.Fatalf("error writing to send pipe: %s\n", err)
		 }
		 err := server.Send.Close()
		 if err != nil {
			 fmt.Printf("failed to close send pipe: %s\n", err)
		 }
	}()

	_, err = io.Copy(os.Stdout, server.Recieve)
	if err != nil {
		log.Fatalf("error copying to stdout: %s\n", err)
	}
}

func client(args ...string) {
	if len(args) != 1 {
		log.Fatalf("USAGE: %s client <ip>\n", os.Args[0])
	}
	ip := parseIp(args[0])

	client := lib.NewClient(ip)

	go func() {
		_, err := io.Copy(client.Send, os.Stdin)
		if err != nil {
			log.Fatalf("error writing to send pipe: %s\n", err)
		}

		err = client.Send.Close()
		if err != nil {
			fmt.Printf("error closing send pipe: %s\n", err)
		}
	}()

	_, err := io.Copy(os.Stdout, client.Receive)
	if err != nil {
		fmt.Printf("error copying to stdout: %s\n", err)
	}
}

func parseIp(s string) [4]byte {
	var ip [4]byte
	split := strings.Split(s, ".")
	if len(split) != 4 {
		log.Fatalf("USAGE: %s client <ip>\n", os.Args[0])
	}

	for i := 0; i < 4; i++ {
		d, err := strconv.Atoi(split[i])
		if err != nil {
			fmt.Printf("failed to parse ip address: %s\n", err)
		}
		ip[i] = byte(d)
	}
	return ip
}
