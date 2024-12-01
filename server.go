package memcached

import (
	"bytes"
	"fmt"
	"io"
	"log"
	store "memcached/internal"
	"net"
)

type ConnectionHandler interface {
	Handle(conn net.Conn, w io.Writer) error
}

type ConnectionHandlerFunc func(net.Conn, io.Writer) error

func (c ConnectionHandlerFunc) Handle(conn net.Conn, w io.Writer) error {
	return c(conn, w)
}

type TCPServer interface {
	Start(handler ConnectionHandler) error
}

type MemcachedServer struct {
	address string
	port    int
	store   store.Store
}

func NewMemcachedServer(address string, port int, store store.Store) (*MemcachedServer, error) {
	return &MemcachedServer{address: address, port: port, store: store}, nil
}

func (m *MemcachedServer) Start(handler ConnectionHandler, w io.Writer) error {
	log.Printf("starting server on %s:%d", m.address, m.port)

	networkAddress := fmt.Sprintf("%s:%d", m.address, m.port)
	ln, err := net.Listen("tcp", networkAddress)
	if err != nil {
		log.Fatalf("an error occurred while setting up Memcached server: %s", err)
		return err
	}

	log.Printf("server running at: %v", ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("an error occurred while setting up Memcached server: %s", err)
			return err
		}
		go handler.Handle(conn, w)
	}
}

func SimpleHandlerFunc(conn net.Conn, w io.Writer) error {
	defer conn.Close()
	n, err := io.Copy(w, conn)

	if err != nil {
		if err != io.EOF {
			log.Fatalf("an error occurred while reading data: %s", err)
			return err
		}
	}

	log.Printf("%d bytes received", n)
	log.Printf("connection to %v closed.\n", conn.RemoteAddr().String())
	return nil
}

type StubConnectionHandler struct {
	receivedDataChan chan []byte
}

func NewStubConnectionHandler(r chan []byte) (*StubConnectionHandler, error) {
	return &StubConnectionHandler{receivedDataChan: r}, nil
}

func (s *StubConnectionHandler) Handle(conn net.Conn, w io.Writer) error {
	defer conn.Close()
	var buf bytes.Buffer
	_, err := io.Copy(&buf, conn)

	if err != nil {
		if err != io.EOF {
			log.Fatalf("an error occurred while reading data: %s", err)
			return err
		}
	}

	// send received data over channel
	s.receivedDataChan <- buf.Bytes()

	return nil
}
