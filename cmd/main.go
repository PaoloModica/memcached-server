package main

import (
	"bufio"
	"flag"
	"log"
	memcached "memcached"
	"os"
)

func main() {
	// set default server address and port
	var address string
	var port int

	flag.StringVar(&address, "a", "127.0.0.1", "server address")
	flag.IntVar(&port, "p", 11211, "server port")
	flag.Parse()

	// TODO - add store

	server, err := memcached.NewMemcachedServer(address, port)

	if err != nil {
		log.Fatalf("an error occurred while setting up memcached server: %s", err)
	}

	connectionHandler := memcached.ConnectionHandlerFunc(memcached.SimpleHandlerFunc)
	out := bufio.NewWriter(os.Stdout)

	go server.Start(connectionHandler, out)
}
