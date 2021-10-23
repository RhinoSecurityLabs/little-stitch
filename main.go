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
		if len(flag.Args()) < 2 {
			log.Fatalf("USAGE: %s client <ip>\n", os.Args[0])
		}
		ip := parseIp(flag.Args()[1])

		conn, wg := lib.NewClient(ip)
		_, err := io.WriteString(conn, "HELL")
		if err != nil {
			log.Fatalf("failed to write to connection: %s\n", err)
		}

		err = conn.Close()
		if err != nil {
			log.Println(err)
		}

		wg.Wait()
	} else if (len(os.Args) >= 1) && (flag.Args()[0] == "server") {
		r, err := lib.NewReceiver()
		if err != nil {
			log.Fatalf("failed to write to connection: %s\n", err)
		}

		_, err = io.Copy(os.Stdout, r.Out)
		if err != nil {
			log.Fatalln("error copying output to stdout")
		}
	} else {
		help()
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
