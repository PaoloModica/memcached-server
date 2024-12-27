package main

import (
	"flag"
	"log"
	app "memcached/internal/app"
	store "memcached/internal/store"
)

func main() {
	// set default server address and port
	var address string
	var port int

	flag.StringVar(&address, "a", "0.0.0.0", "server address")
	flag.IntVar(&port, "p", 11211, "server port")
	flag.Parse()

	store := store.NewInMemoryStore()

	server, err := app.NewMemcachedServer(address, port, store)

	if err != nil {
		log.Fatalf("an error occurred while setting up memcached server: %s", err)
	}

	server.Start()
}
