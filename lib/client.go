package lib

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"syscall"
	"time"
)

const TIMEOUT = time.Second
const SEND_BASE_PORT = 11000
const RECEIVE_BASE_PORT = 11010

func NewClient(addr [4]byte) *Client {
	sendR, sendW := io.Pipe()
	receiveR, receiveW := io.Pipe()
	client := &Client{
		Addr: addr,
		Send: sendW,
		sendReader: sendR,
		Receive: receiveR,
		receiveWriter: receiveW,
		wg: &sync.WaitGroup{},
	}
	client.start()
	return client
}

type Client struct {
	Addr          [4]byte
	Send          *io.PipeWriter
	sendReader    *io.PipeReader
	Receive       *io.PipeReader
	receiveWriter *io.PipeWriter
	wg            *sync.WaitGroup
}

func (c *Client) start() {
	c.wg.Add(1)
	go c.sendWorker()

	c.wg.Add(1)
	go c.receiveWorker()
}

func (c *Client) sendWorker() {
	var err error
	defer func() {
		if err != nil {
			fmt.Printf("error in send worker: %s\n", err)
		}
		c.wg.Done()
	}()

	b, err := ioutil.ReadAll(c.sendReader)
	if err != nil {
		return
    }

	// Iterate over each byte in the string.
	for _, b := range b {
		 // Iterate over each bit in the byte.
		 for bit := 1; bit <= 8; bit++ {
			 // We want to check if the bit we are checking is set for the current byte. To do this we can shift it
			 // right n-1, bitwise and it with 1 to clear bits on the left, then check if it equals 1.
			 if (b >> (bit-1) & 1) == 1 {
				 _, err := Ping(&syscall.SockaddrInet4{Addr: c.Addr, Port: SEND_BASE_PORT + bit})
				 if err != nil {
					  fmt.Printf("[ERROR] Recieved error when sending bit %d: %s\n", bit, err)
				 }
			 }
		 }

		// Send clock ping, notifies the server we finished the last byte.
		_, err = Ping(&syscall.SockaddrInet4{Addr: c.Addr, Port: SEND_BASE_PORT})
		if err != nil {
			fmt.Printf("[ERROR] Recieved error when sending clock ping: %s\n", err)
		}
		time.Sleep(1)

	}
	return
}

func (c *Client) receiveWorker() {
	var err error
	defer func() {
		if err != nil {
			fmt.Printf("error in send worker: %s\n", err)
		}
		c.wg.Done()
	}()

	// Iterate over each byte until the clock port stops responding
	for {
		err := c.receiveClockPing()
		if err != nil {
			return
		}

		var b byte
		// Iterate over each bit in the byte.
		for bit := 1; bit <= 8; bit++ {
			open, err := Ping(&syscall.SockaddrInet4{Addr: c.Addr, Port: RECEIVE_BASE_PORT + bit})
			if err != nil {
				fmt.Printf("[ERROR] Recieved error when sending bit %d: %s\n", bit, err)
			}
			if open {
				b |= 1 << (bit - 1)
			}
		}
		written, err := bytes.NewBuffer([]byte{b}).WriteTo(c.receiveWriter)
		if err != nil || written != 1 {
			fmt.Printf("failed to write %v to pipe: %s", b, err)
		}
	}
}

func (c *Client) receiveClockPing() error {
	checks := 10
	var err error
	var open bool

	// Try sending the Clock ping 10 times, if the port isn't open by then something went wrong.
	for i := 1; i <= checks; i++ {
		// Server will close this port when it is ready to send data.
		open, err = Ping(&syscall.SockaddrInet4{Addr: c.Addr, Port: RECEIVE_BASE_PORT})
		if err != nil || !open {
			fmt.Printf("error sending clock ping (# %d of %d): %s\n", i, checks, err)
			time.Sleep(time.Second)
		}
	}

	if err != nil || !open {
		err = fmt.Errorf("failed to establish clock connection after %d times\n", checks)
	}
	return err
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
			return false, fmt.Errorf("sendWorker rejected")
		} else if time.Now().After(start.Add(TIMEOUT)) {
			return false, fmt.Errorf("sendWorker timedout")
		} else if err.Error() == "socket is not connected" {
			time.Sleep(TIMEOUT / 20)
			_, err = syscall.Getpeername(sock)
		}
	}
	return true, nil
}
