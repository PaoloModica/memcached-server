package main

import (
	"flag"
	"log"
	memcached "memcached"
	store "memcached/internal"
)

func main() {
	// set default server address and port
	var address string
	var port int

	flag.StringVar(&address, "a", "127.0.0.1", "server address")
	flag.IntVar(&port, "p", 11211, "server port")
	flag.Parse()

	store := store.NewInMemoryStore()

	server, err := memcached.NewMemcachedServer(address, port, store)

	if err != nil {
		log.Fatalf("an error occurred while setting up memcached server: %s", err)
	}

	server.Start()
}
