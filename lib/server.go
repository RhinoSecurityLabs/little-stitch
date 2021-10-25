package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
)

type Server struct {
	Recieve       *io.PipeReader
	recieveWriter *io.PipeWriter
	Send       	  *io.PipeWriter
	sendReader    *io.PipeReader
	mutex         *sync.Mutex
	b             byte
	ready         chan int
}

func NewServer() (*Server, error) {
	recieveRPipe, recieveWPipe := io.Pipe()
	sendRPipe, sendWPipe := io.Pipe()

	server := &Server{
		Recieve: recieveRPipe,
		recieveWriter: recieveWPipe,
		Send: sendWPipe,
		sendReader: sendRPipe,
		mutex: &sync.Mutex{},
		b:     byte(0),
		ready: make(chan int, 1),
	}

	err := server.receiveWorker()
	if err != nil {
		return nil, fmt.Errorf("receive worker: %s\n", err)
	}

	server.sendWorker()

	return server, nil
}


func (s *Server) receiveWorker() (err error) {
	err = s.handleSendConn(SEND_BASE_PORT, func() {
		if s.b == byte(0) {
			fmt.Println("r.b is nil")
			return
		}

		_, err := s.recieveWriter.Write([]byte{s.b})
		if err != nil {
			 fmt.Printf("failed writing recieved byte to pipe: %s\n", err)
		}
		s.b = byte(0) // Clear byte before next iteration
	})
	if err != nil {
		return fmt.Errorf("clock listener: %s\n", err)
	}

	for i := 1; i <= 8; i++ {
		err := s.handleSendConn(SEND_BASE_PORT + i, func(bit int) func() {
			return func() {
				// We want the nth bit from right set, to do that we can take 1 and shift it left n-1.
				s.b |= byte(1 << (bit-1)) // Is this atomic?
			}
		}(i))
		if err != nil {
			  log.Fatalf("bit %d: %s\n", i, err)
		}
	}
	return nil
}

func (s *Server) waitForConn(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return fmt.Errorf("unable to listen on port %d: %s\n", port, err)
	}

	conn, err := ln.Accept()
	if err != nil {
		return fmt.Errorf("error accepting connection: %s\n", err)
	}

	err = conn.Close()
	if err != nil {
		 return fmt.Errorf("closing connection on port %d: %s\n", port, err)
	}

	err = ln.Close()
	if err != nil {
		return fmt.Errorf("closing port %d: %s\n", port, err)
	}

	return nil
}

func (s *Server) handleSendConn(port int, f func()) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return fmt.Errorf("unable to listen on port %d: %s\n", port, err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("error accepting connection:", err)
			}

			s.mutex.Lock()
			f()
			s.mutex.Unlock()

			err = conn.Close()
			if err != nil {
				fmt.Println("error closing port:", err)
			}
		}
	}()
	return nil
}


func (s *Server) sendWorker() {
	shouldOpen := map[int]chan int{}

	for i := 1; i <= 8; i++ {
		shouldOpen[i] = make(chan int, 1)
		s.handleSend(RECEIVE_BASE_PORT + i, func(bit int, shouldOpen chan int) func() {
			return func() {
				<-shouldOpen
			}
		}(i, shouldOpen[i]))
	}

	go func() {
		 buf, err := ioutil.ReadAll(s.sendReader)
		 if err != nil {
			 fmt.Printf("reading from send pipe: %s\n", err)
		 }
		 for _, b := range buf {
			 // Opening the clock port means we're ready for the client to check the bit ports.

			 bCopy := b
			 for bit := 1; bit <= 8; bit++ {
				 isSet := ((bCopy >> (bit - 1)) & 1) == byte(1)
				 if isSet {
					  shouldOpen[bit] <- 1
				 }
			 }

			 // Opening the start clock port means we're ready for the client to check the bit ports.
			 err = s.waitForConn(RECEIVE_BASE_PORT)
			 if err != nil {
				 fmt.Printf("server send clock start: %s\n", err)
			 }

			 // A connection on the end clock port means the client has finished iterating the ports.
			 err = s.waitForConn(RECEIVE_BASE_PORT + 9)
			 if err != nil {
				  fmt.Printf("server send clock end: %s\n", err)
			 }
		 }
	}()
}

func (s *Server) handleSend(port int, f func()) {
	go func() {
		for {
			f()

			ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
			if err != nil {
				fmt.Printf("unable to listen on port %d: %s\n", port, err)
			}

			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("error accepting connection:", err)
			}

			err = conn.Close()
			if err != nil {
				fmt.Printf("error closing connection on port %d: %s\n", port, err)
			}

			err = ln.Close()
			if err != nil {
				fmt.Printf("error closing port %d: %s\n", port, err)
			}
		}
	}()
}
