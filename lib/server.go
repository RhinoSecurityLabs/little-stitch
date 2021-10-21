package lib

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

func NewReceiver() (*Reciever, error) {
	rPipe, wPipe := io.Pipe()
	r := &Reciever{
		Out:   rPipe,
		data:  wPipe,
		mutex: &sync.Mutex{},
		b:     byte(0),
		ready: make(chan int, 1),
	}
	for i := 1; i <= 8; i++ {
		go func(i int) {
			if err := r.handleConnection(i); err != nil {
				log.Fatalf("handling connection failed: %s\n", err)
			}
		}(i)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", BASE_PORT))
	if err != nil {
		return nil, fmt.Errorf("unable to listen on port %d: %s\n", BASE_PORT, err)
	}

	// Clock port, when we receive a connection on this port it signifies the start of the next byte.
	go func() {
		// First connection can be ignored.
		//conn, _ := ln.Accept()
		//conn.Close()

		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Printf("error accepting connection: %s\n", err)
			}
			if r.b == byte(0) {
				fmt.Println("r.b is nil")
				continue
			}
			fmt.Println("got: ", string(r.b))

			r.mutex.Lock()
			for {
				write, err := r.data.Write([]byte{r.b})
				if err != nil {
					fmt.Printf("failed writing recieved byte to pipe: %s\n", err)
				}
				r.b = byte(0) // Clear byte before next iteration
				if write >= 1 {
					break
				}
			}
			r.mutex.Unlock()
			err = conn.Close()
			if err != nil {
				fmt.Printf("failled to close connection: %s\n", err)
			}
		}
	}()
	return r, nil
}

type Reciever struct {
	Out *io.PipeReader
	data *io.PipeWriter
	mutex *sync.Mutex
	b byte
	ready chan int
}

func (r *Reciever) handleConnection(bit int) error {
	port := BASE_PORT + bit
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return fmt.Errorf("unable to listen on port %d: %s\n", BASE_PORT, err)
	}
	fmt.Printf("listener for bit %d setup on port %d\n", bit, port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error accepting connection:", err)
		}
		fmt.Println("got ping for bit", bit)

		r.mutex.Lock()
		r.b |= byte(bit) // Is this atomic?
		r.mutex.Unlock()

		err = conn.Close()
		if err != nil {
			fmt.Println("error closing port:", err)
		}
	}
}
