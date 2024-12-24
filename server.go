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
	"time"
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
	log.Printf("managing connection %v", conn.RemoteAddr())
	defer conn.Close()

	scanner := bufio.NewScanner(bufio.NewReader(conn))

	var command string
	var cmdComponents [6]string
	var keyToAdd string
	var noReply string
	var keyEntry store.MapEntry

	for scanner.Scan() {
		command = scanner.Text()
		components := strings.Split(command, " ")
		copy(cmdComponents[:], components)

		if len(cmdComponents) < 1 {
			log.Print("not enough arguments for the command")
			continue
		}

		switch cmdComponents[0] {
		case "add":
			log.Println("processing ADD command")
			keyToAdd = cmdComponents[1]

			// check whether the key already exists in store
			keyVal, err := m.store.Get(keyToAdd)
			if keyVal == nil && err != nil {
				flags, _ := strconv.Atoi(cmdComponents[2])
				expiration, _ := strconv.Atoi(cmdComponents[3])
				noReply = cmdComponents[5]

				keyEntry = store.MapEntry{Flags: uint16(flags), Expiration: time.Duration(expiration)}
			} else {
				log.Printf("value already existing for key %s, nothing to add", keyToAdd)
				conn.Write([]byte("NOT_STORED\r\n"))
				keyToAdd = ""
			}
		case "get":
			log.Println("processing GET command")
			keyVal, err := m.store.Get(cmdComponents[1])
			if err != nil {
				log.Printf("an error occurred while fetching key from store: %s", err.Error())
				conn.Write([]byte("END\r\n"))
				continue
			}
			var output string
			if !keyVal.IsExpired() {
				output = fmt.Sprintf("VALUE %s %d %d\r\n%s\r\nEND\r\n", cmdComponents[1], keyVal.Flags, len(keyVal.Data), string(keyVal.Data))
			} else {
				m.store.Remove(cmdComponents[1])
				output = "END\r\n"
			}
			conn.Write([]byte(output))
		case "replace":
			log.Println("processing REPLACE command")
			keyToAdd = cmdComponents[1]

			// check whether the key already exists in store
			keyVal, err := m.store.Get(keyToAdd)
			if keyVal == nil && err != nil {
				log.Printf("no value found for key %s, nothing to replace", keyToAdd)
				conn.Write([]byte("NOT_STORED\r\n"))
				keyToAdd = ""
			} else {
				flags, _ := strconv.Atoi(cmdComponents[2])
				expiration, _ := strconv.Atoi(cmdComponents[3])
				noReply = cmdComponents[5]

				keyEntry = store.MapEntry{Flags: uint16(flags), Expiration: time.Duration(expiration)}
			}
		case "set":
			log.Println("processing SET command")
			keyToAdd = cmdComponents[1]
			flags, _ := strconv.Atoi(cmdComponents[2])
			expiration, _ := strconv.Atoi(cmdComponents[3])
			noReply = cmdComponents[5]

			keyEntry = store.MapEntry{Flags: uint16(flags), Expiration: time.Duration(expiration)}
		default:
			if keyToAdd != "" {
				log.Printf("store %s in key %s", cmdComponents[0], keyToAdd)
				m.store.Add(keyToAdd, []byte(cmdComponents[0]), keyEntry.Flags, keyEntry.Expiration)
				if noReply == "" {
					conn.Write([]byte("STORED\r\n"))
				} else {
					conn.Write([]byte("\r\n"))
				}
				// reset keyToAdd and keyEntry
				keyToAdd = ""
				noReply = ""
				keyEntry = store.MapEntry{}
			} else {
				log.Print("command not recognised")
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
