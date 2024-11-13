package memcached_test

import (
	"bytes"
	"fmt"
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
	testServerPort := 11211
	receivedDataChan := make(chan []byte)
	stubConnectionHandler, _ := memcached.NewStubConnectionHandler(receivedDataChan)
	server, _ := memcached.NewMemcachedServer(testServerAddress, testServerPort)

	t.Run("server accept incoming request", func(t *testing.T) {
		messageData := "test data"
		buf := bytes.Buffer{}
		// dial a connection to the server to send data
		serverAddress := fmt.Sprintf("%s:%d", testServerAddress, testServerPort)

		go server.Start(stubConnectionHandler, &buf)

		go dialConnection(t, serverAddress, messageData, 5)

		select {
		case receivedData := <-receivedDataChan:
			if string(receivedData) != messageData {
				t.Errorf("expected to receive %s, got %s", messageData, receivedData)
			}
		case <-time.After(30 * time.Second):
			t.Fatalf("TCP server test time out")
		}
	})
}
