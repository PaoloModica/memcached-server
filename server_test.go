// ToDO - review test after store addition to MemcachedServer
package memcached_test

import (
	"bytes"
	"fmt"
	"log"
	memcached "memcached"
	"net"
	"testing"
	"time"
)

func dialConnection(t *testing.T, address string, message string, retry int) error {
	t.Helper()

	for i := 0; i < retry; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			continue
		}

		defer conn.Close()

		n, err := conn.Write([]byte(message))

		if err != nil || n == 0 {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func TestServer(t *testing.T) {
	testServerAddress := "127.0.0.1"
	testServerPort := 11212
	receivedDataChan := make(chan []byte)
	stubConnectionHandler, _ := memcached.NewStubConnectionHandler(receivedDataChan)
	server, _ := memcached.NewMemcachedServer(testServerAddress, testServerPort)
	serverAddress := fmt.Sprintf("%s:%d", testServerAddress, testServerPort)
	buf := bytes.Buffer{}

	go server.Start(stubConnectionHandler, &buf)

	t.Run("server accept incoming request", func(t *testing.T) {
		messageData := "test data"

		// dial a connection to the server to send data
		go dialConnection(t, serverAddress, messageData, 5)

		select {
		case receivedData := <-receivedDataChan:
			log.Printf("received message: %s", receivedData)
			if string(receivedData) != messageData {
				t.Errorf("expected to receive %s, got %s", messageData, receivedData)
			}
		case <-time.After(30 * time.Second):
			t.Fatalf("TCP server test time out")
		}
	})
	t.Run("GET command returns key", func(t *testing.T) {
		command := "get test\r\n"
		expectedData := "VALUE test 1 1"

		// dial a connection to the server to send data
		go dialConnection(t, serverAddress, command, 5)

		select {
		case receivedData := <-receivedDataChan:
			log.Printf("received message: %s", receivedData)
			if string(receivedData) != expectedData {
				t.Errorf("expected to receive %s, got %s", expectedData, receivedData)
			}
		case <-time.After(30 * time.Second):
			t.Fatalf("TCP server test time out")
		}
	})

	t.Run("SET command stores new key value pair", func(t *testing.T) {
		command := "set test 0 0 4\r\n1234\r\n"
		expectedData := "STORED"

		// dial a connection to the server to send data
		go dialConnection(t, serverAddress, command, 5)

		select {
		case receivedData := <-receivedDataChan:
			log.Printf("received message: %s", receivedData)
			if string(receivedData) != expectedData {
				t.Errorf("expected to receive %s, got %s", expectedData, receivedData)
			}
		case <-time.After(30 * time.Second):
			t.Fatalf("TCP server test time out")
		}
	})
}
