package lib

import (
	"fmt"
	"io"
	"syscall"
	"time"
)

const TIMEOUT = time.Second
const BASE_PORT = 60000

func NewConnection(addr [4]byte) *io.PipeWriter {
	r, w := io.Pipe()
	go connection(addr, r)
	return w
}

func connection(addr [4]byte, in *io.PipeReader) error {
	for {
		var b []byte
		_, err := in.Read(b)
		if err != nil {
			return err
		}

		// Send clock ping, notifies the server we are starting the next byte.
		_, err = Ping(&syscall.SockaddrInet4{Addr: addr, Port: BASE_PORT})
		if err != nil {
			fmt.Printf("[ERROR] Recieved error when sending clock ping: %s\n", err)
		}
		for _, b := range b {
			for i := 1; i <= 8; i++ {
				err := sendBit(addr, b, i)
				if err != nil {
					fmt.Printf("[ERROR] Recieved error when sending bit %d: %s\n", i, err)
				}
			}
		}
	}
}

func sendBit(addr [4]byte, b byte, bit int) error {
	if (int(b) & bit) == 1 {
		_, err := Ping(&syscall.SockaddrInet4{Addr: addr, Port: BASE_PORT + bit})
		return err
	}
	return nil
}

func Ping(addr syscall.Sockaddr) (bool, error) {
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return false, fmt.Errorf("creating socket: %w\n", err)
	}
	defer func(sock int) {
		err := syscall.Close(sock)
		if err != nil {
			fmt.Printf("Failed to close socket %d\n", sock)
		}
	}(sock)

	//fmt.Println("socket resp: ", sock)

	err = syscall.Connect(sock, addr)
	for err == syscall.EINTR {
		err = syscall.Connect(sock, addr)
	}

	start := time.Now()
	_, err = syscall.Getpeername(sock)

	for err != nil {
		if err.Error() == "invalid argument" {
			return false, fmt.Errorf("connection rejected")
		} else if time.Now().After(start.Add(TIMEOUT)) {
			return false, fmt.Errorf("connection timedout")
		} else if err.Error() == "socket is not connected" {
			time.Sleep(TIMEOUT / 20)
			_, err = syscall.Getpeername(sock)
		}
	}
	return true, nil
}
