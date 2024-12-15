package memcached

import (
	"bufio"
	"fmt"
	"io"
	"log"
	store "memcached/internal"
	"net"
	"strconv"
	"strings"
)

type CommandNotValidError string

func (e CommandNotValidError) Error() string {
	return string(e)
}

type TCPServer interface {
	Start(w io.Writer) error
}

type MemcachedServer struct {
	address string
	port    int
	store   store.Store
}

func NewMemcachedServer(address string, port int, store store.Store) (*MemcachedServer, error) {
	return &MemcachedServer{address: address, port: port, store: store}, nil
}

func (m *MemcachedServer) handleConnection(conn net.Conn) {
	log.Printf("Managing connection %v", conn.RemoteAddr())
	defer conn.Close()

	scanner := bufio.NewScanner(bufio.NewReader(conn))

	var command string
	var keyToAdd string
	var keyEntry store.MapEntry

	for scanner.Scan() {
		command = scanner.Text()

		cmdComponents := strings.Split(command, " ")

		log.Printf("Command components: %v", cmdComponents)

		if len(cmdComponents) < 1 {
			log.Print("not enough arguments for the command")
			continue
		}

		switch cmdComponents[0] {
		case "set":
			log.Println("Processing SET command")
			keyToAdd = cmdComponents[1]
			flags, _ := strconv.Atoi(cmdComponents[2])

			keyEntry = store.MapEntry{Flags: uint16(flags)}
		case "get":
			log.Println("Processing GET command")
			keyVal, err := m.store.Get(cmdComponents[1])
			if err != nil {
				conn.Write([]byte(err.Error()))
				continue
			}
			output := fmt.Sprintf("VALUE %s %d %d\r\n%s\r\nEND\r\n", cmdComponents[1], keyVal.Flags, len(keyVal.Data), string(keyVal.Data))
			conn.Write([]byte(output))
		default:
			if keyToAdd != "" {
				log.Printf("Store %s in key %s", cmdComponents[0], keyToAdd)
				m.store.Add(keyToAdd, []byte(cmdComponents[0]), keyEntry.Flags)
				conn.Write([]byte("STORED\r\n"))
				// reset keyToAdd and keyEntry
				keyToAdd = ""
				keyEntry = store.MapEntry{}
			} else {
				log.Print("Command not recognised")
				conn.Write([]byte(fmt.Sprintf("%s command not recognised\r\n", cmdComponents[0])))
			}
		}
	}
}

func (m *MemcachedServer) Start() error {
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
		go m.handleConnection(conn)
	}
}
