package memcached_test

import (
	"bufio"
	"errors"
	"fmt"
	memcached "memcached"
	store "memcached/internal"
	"net"
	"testing"
	"time"
)

func dialConnection(t *testing.T, address string, message string, retry int) (net.Conn, error) {
	t.Helper()

	for i := 0; i < retry; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			continue
		}

		time.Sleep(200 * time.Millisecond)
		n, err := conn.Write([]byte(message))

		if err != nil || n == 0 {
			return nil, err
		} else {
			return conn, nil
		}
	}
	return nil, errors.New("failed dialing connection")
}

func TestServer(t *testing.T) {
	testServerAddress := "127.0.0.1"
	testServerPort := 11212
	stubStore := store.NewInMemoryStore()
	stubStore.Add("test", []byte("1234"), 0)
	server, _ := memcached.NewMemcachedServer(testServerAddress, testServerPort, stubStore)
	serverAddress := fmt.Sprintf("%s:%d", testServerAddress, testServerPort)

	go server.Start()
	time.Sleep(100 * time.Millisecond)

	t.Run("server accept incoming request", func(t *testing.T) {
		messageData := "hello world\n"
		expectedData := "hello command not recognised"

		// dial a connection to the server to send data
		conn, err := dialConnection(t, serverAddress, messageData, 5)

		if err != nil {
			t.Errorf("expected connection to get established, got error %s", err.Error())
		}

		defer conn.Close()

		connectionScanner := bufio.NewScanner(bufio.NewReader(conn))

		var receivedData string
		expectedReceivedDataLines := 1
		i := 0
		fmt.Print("scanning for data to be received back")
		for connectionScanner.Scan() {

			receivedData += connectionScanner.Text()

			if len(receivedData) > 0 {
				i += 1
			}

			if i == expectedReceivedDataLines {
				break
			}
		}

		if receivedData != expectedData {
			t.Errorf("expected to receive %s, got %s", expectedData, receivedData)
		}
	})
	t.Run("GET command returns key", func(t *testing.T) {
		command := "get test\r\n"
		expectedData := "VALUE test 0 4\r\n1234\r\nEND\r\n"

		// dial a connection to the server to send data
		conn, err := dialConnection(t, serverAddress, command, 5)

		if err != nil {
			t.Errorf("expected connection to get established, got error %s", err.Error())
		}

		defer conn.Close()

		connectionScanner := bufio.NewScanner(bufio.NewReader(conn))

		var receivedData string
		expectedReceivedDataLines := 3
		i := 0
		fmt.Print("scanning for data to be received back")
		for connectionScanner.Scan() {

			receivedData += connectionScanner.Text()

			if len(receivedData) > 0 {
				i += 1
				receivedData += "\r\n"
			}

			if i == expectedReceivedDataLines {
				break
			}
		}

		if receivedData != expectedData {
			t.Errorf("expected to receive %s, got %s", expectedData, receivedData)
		}
	})

	t.Run("SET command stores new key value pair", func(t *testing.T) {
		command := "set test2 0 0 4\r\n9876\r\n"
		expectedData := "STORED\r\n"

		// dial a connection to the server to send data
		conn, err := dialConnection(t, serverAddress, command, 5)

		if err != nil {
			t.Errorf("expected connection to get established, got error %s", err.Error())
		}

		defer conn.Close()

		connectionScanner := bufio.NewScanner(bufio.NewReader(conn))

		var receivedData string
		expectedReceivedDataLines := 1
		i := 0
		fmt.Print("scanning for data to be received back")
		for connectionScanner.Scan() {

			receivedData += connectionScanner.Text()

			if len(receivedData) > 0 {
				i += 1
				receivedData += "\r\n"
			}

			if i == expectedReceivedDataLines {
				break
			}
		}

		if receivedData != expectedData {
			t.Errorf("expected to receive %s, got %s", expectedData, receivedData)
		}
	})
}
