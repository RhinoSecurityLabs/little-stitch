package lib

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)


const SendBasePort = 11100
const ReceiveBasePort = 11200
const ReceiveSendWidth = 2

const TIMEOUT = time.Second * 3

func NewClient(addr []byte) *Client {
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
	Addr          []byte
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

	for {
		b := make([]byte, 8)
		_, err := c.sendReader.Read(b)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("error reading from send pipe: %s\n", err)
			continue
		}

		// Iterate over each byte in the string.
		for _, b := range b {
			// Iterate over each bit in the byte.
			for bit := 1; bit <= 8; bit++ {
				// We want to check if the bit we are checking is set for the current byte. To do this we can shift it
				// right n-1, bitwise and it with 1 to clear bits on the left, then check if it equals 1.
				if (b >> (bit - 1) & 1) == 1 {
					// TODO: Debug logging of errors returned here.

					Ping(c.Addr, SendBasePort+ bit)
				}
			}

			// Send clock ping, notifies the server we finished the last byte.
			// TODO: Debug logging of errors returned here.
			Ping(c.Addr, SendBasePort)
		}
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

		b := map[int]byte{}
		m := &sync.Mutex{}
		wg := &sync.WaitGroup{}

		for o := 0; o < ReceiveSendWidth; o++ {
			b[o] = byte(0)

			// Iterate over each bit in the byte.
			for bit := 1; bit <= 8; bit++ {
				// TODO: Debug logging of errors returned here.
				wg.Add(1)
				go func(b *map[int]byte, bit, o int) {
					open, _ := Ping(c.Addr, ReceiveBasePort + (o * 10) + bit)
					if open {
						m.Lock()
						(*b)[o] |= 1 << (bit - 1)
						m.Unlock()
					}
					wg.Done()
				}(&b, bit, o)
			}
		}
		wg.Wait()

		for i := 0; i < ReceiveSendWidth; i++ {
			written, err := bytes.NewBuffer([]byte{b[i]}).WriteTo(c.receiveWriter)
			if err != nil || written != 1 {
				fmt.Printf("failed to write %v to pipe: %s", b, err)
			}
		}

		// Signals to the server we're finished processing this byte.
		_, err = Ping(c.Addr, ReceiveBasePort+ 9)
		if err != nil {
			fmt.Printf("error when sending clock end ping: %s\n", err)
		}
	}
}

func (c *Client) receiveClockPing() error {
	checks := 9
	var err error
	var open bool

	wait := time.Millisecond
	// Try sending the Clock ping 10 times, if the port isn't open by then something went wrong.
	for i := 1; i <= checks; i++ {
		// TODO: Debug logging of retries.
		open, err = Ping(c.Addr, ReceiveBasePort)
		if err != nil || !open {
			time.Sleep(wait)
			wait *= 3
		} else {
			break
		}
	}

	if err != nil || !open {
		err = fmt.Errorf("failed to establish clock connection after %d times\n", checks)
	}
	return err
}

func (c *Client) Wait() (err error) {
	c.wg.Wait()

	err = c.Receive.Close()
	if err != nil {
		return
	}

	err = c.receiveWriter.Close()
	if err != nil {
		return
	}
	return
}

func Ping(ip []byte, port int) (bool, error) {
	addr := net.TCPAddr{IP: ip, Port: port}

	d := net.Dialer{Timeout: TIMEOUT}
	conn, err := d.Dial("tcp", addr.String())

	if err != nil {
		if strings.Contains(err.Error(), "connect: connection refused") {
			return false, err
		} else {
			return false, err
		}
	} else {
		err := conn.Close()
		if err != nil {
			fmt.Println("error closing connection")
		}
		return true, nil
	}
}

//func Ping(ip []byte, port int) (bool, error) {
//	var _ip [4]byte
//	copy(_ip[:], ip[:4])
//	addr := &syscall.SockaddrInet4{Addr: _ip, Port: port}
//
//	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
//	if err != nil {
//		return false, fmt.Errorf("creating socket: %w\n", err)
//	}
//
//	err = syscall.SetsockoptInt(sock, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
//	if err != nil {
//		return false, fmt.Errorf("setting SO_REUSEPORT on socket: %s", err)
//	}
//
//	err = syscall.Connect(sock, addr)
//	for err == syscall.EINTR {
//		err = syscall.Connect(sock, addr)
//	}
//
//	_, err = syscall.Getpeername(sock)
//
//	for err != nil {
//		if err.Error() == "invalid argument" {
//			return false, fmt.Errorf("sendWorker rejected")
//		} else if err.Error() == "socket is not connected" {
//			time.Sleep(TIMEOUT)
//			_, err = syscall.Getpeername(sock)
//		}
//	}
//
//	go func() {
//		err := syscall.Close(sock)
//		if err != nil {
//			fmt.Printf("Failed to close socket %d\n", sock)
//		}
//	}()
//	return true, nil
//}
