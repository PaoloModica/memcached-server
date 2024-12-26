package memcached_test

import (
	"bufio"
	"errors"
	"fmt"
	memcached "memcached"
	store "memcached/internal"
	"net"
	"strings"
	"testing"
	"time"
)

type testCase struct {
	Id             string
	Command        string
	ExpectedResult string
}

type appendPrependTestCase struct {
	testCase
	ExpectedKeyVal string
}

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

func dialConnectionAndCheckReceivedData(t *testing.T, addr string, tc testCase) (receivedData string) {
	t.Helper()
	// dial a connection to the server to send data
	conn, err := dialConnection(t, addr, tc.Command, 5)

	if err != nil {
		t.Errorf("expected connection to get established, got error %s", err.Error())
	}

	defer conn.Close()

	connectionScanner := bufio.NewScanner(bufio.NewReader(conn))

	expectedReceivedDataLines := strings.Count(tc.ExpectedResult, "\r\n")
	i := 0
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
	return
}

func assertKeyContent(t *testing.T, store store.Store, key string, exp string) {
	t.Helper()

	keyVal, _ := store.Get(key)
	if string(keyVal.Data) != exp {
		t.Errorf("expected key %s to feature %s content, found %s", key, exp, string(keyVal.Data))
	}
}

func TestServer(t *testing.T) {
	testServerAddress := "127.0.0.1"
	testServerPort := 11212
	stubStore := store.NewInMemoryStore()
	stubStore.Add("test1", []byte("1234"), 0, 100) // key which should be valid at test time
	stubStore.Add("test2", []byte("9876"), 0, -1)  // key already expired
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
	t.Run("GET command tests", func(t *testing.T) {
		getCmdTestCases := []testCase{
			{
				Id:             "unknown key return nothing",
				Command:        "get test\r\n",
				ExpectedResult: "END\r\n",
			},
			{
				Id:             "valid key return key value",
				Command:        "get test1\r\n",
				ExpectedResult: "VALUE test1 0 4\r\n1234\r\nEND\r\n",
			},
			{
				Id:             "expired key return nothing",
				Command:        "get test2\r\n",
				ExpectedResult: "END\r\n",
			},
		}

		for _, tc := range getCmdTestCases {
			t.Run(tc.Id, func(t *testing.T) {
				receivedData := dialConnectionAndCheckReceivedData(t, serverAddress, tc)

				if receivedData != tc.ExpectedResult {
					t.Errorf("expected to receive %s, got %s", tc.ExpectedResult, receivedData)
				}
			})
		}
	})
	t.Run("SET command tests", func(t *testing.T) {
		setCmdTestCases := []testCase{
			{
				Id:             "store new key pair, return stored confirmation",
				Command:        "set test3 0 1 4\r\n9876\r\n",
				ExpectedResult: "STORED\r\n",
			},
			{
				Id:             "store new key pair with no reply option, return nothing",
				Command:        "set test4 0 1 4 noreply\r\n09182736\r\n",
				ExpectedResult: "",
			},
		}

		for _, tc := range setCmdTestCases {
			t.Run(tc.Id, func(t *testing.T) {
				receivedData := dialConnectionAndCheckReceivedData(t, serverAddress, tc)

				if receivedData != tc.ExpectedResult {
					t.Errorf("expected to receive %s, got %s", tc.ExpectedResult, receivedData)
				}
			})
		}
	})
	t.Run("ADD command tests", func(t *testing.T) {
		addCmdTestCases := []testCase{
			{
				Id:             "add existing key, return not stored",
				Command:        "add test3 0 1 4\r\n9876\r\n",
				ExpectedResult: "NOT_STORED\r\n",
			},
			{
				Id:             "add new key pair, return stored",
				Command:        "add test5 0 1 4\r\n09182736\r\n",
				ExpectedResult: "STORED\r\n",
			},
			{
				Id:             "add new key pair, no reply, return nothing",
				Command:        "add test6 0 1 4 noreply\r\n09182736\r\n",
				ExpectedResult: "",
			},
		}

		for _, tc := range addCmdTestCases {
			t.Run(tc.Id, func(t *testing.T) {
				receivedData := dialConnectionAndCheckReceivedData(t, serverAddress, tc)

				if receivedData != tc.ExpectedResult {
					t.Errorf("expected to receive %s, got %s", tc.ExpectedResult, receivedData)
				}
			})
		}
	})
	t.Run("REPLACE command tests", func(t *testing.T) {
		replaceCmdTestCases := []testCase{
			{
				Id:             "replace unknown key content, return not stored",
				Command:        "replace test99 0 1 4\r\n9876\r\n",
				ExpectedResult: "NOT_STORED\r\n",
			},
			{
				Id:             "replace existing key content, return stored",
				Command:        "replace test3 0 1 4\r\n54321\r\n",
				ExpectedResult: "STORED\r\n",
			},
			{
				Id:             "replace existing key content noreply, return nothing",
				Command:        "replace test3 0 1 4 noreply\r\012938\r\n",
				ExpectedResult: "",
			},
		}

		for _, tc := range replaceCmdTestCases {
			t.Run(tc.Id, func(t *testing.T) {
				receivedData := dialConnectionAndCheckReceivedData(t, serverAddress, tc)

				if receivedData != tc.ExpectedResult {
					t.Errorf("expected to receive %s, got %s", tc.ExpectedResult, receivedData)
				}
			})
		}
	})
	t.Run("APPEND command tests", func(t *testing.T) {
		appendCmdTestCases := []appendPrependTestCase{
			{
				testCase: testCase{
					Id:             "append data to unknown key content, return not stored",
					Command:        "append test99 0 1 4\r\n9876\r\n",
					ExpectedResult: "NOT_STORED\r\n",
				},
				ExpectedKeyVal: "",
			},
			{
				testCase: testCase{
					Id:             "append data to existing key content, return stored",
					Command:        "append test1 0 1 4\r\nmore\r\n",
					ExpectedResult: "STORED\r\n",
				},
				ExpectedKeyVal: "1234more",
			},
			{
				testCase: testCase{
					Id:             "append data to existing key content noreply, return nothing",
					Command:        "append test1 0 1 4 noreply\r\nmore\r\n",
					ExpectedResult: "",
				},
				ExpectedKeyVal: "1234moremore",
			},
		}

		for _, tc := range appendCmdTestCases {
			t.Run(tc.Id, func(t *testing.T) {
				receivedData := dialConnectionAndCheckReceivedData(t, serverAddress, tc.testCase)

				if receivedData != tc.ExpectedResult {
					t.Errorf("expected to receive %s, got %s", tc.ExpectedResult, receivedData)
				}

				key := strings.Split(tc.Command, " ")[1]

				if tc.ExpectedKeyVal != "" {
					assertKeyContent(t, stubStore, key, tc.ExpectedKeyVal)
				}

			})
		}
	})
	t.Run("PREPEND command tests", func(t *testing.T) {
		prependCmdTestCases := []appendPrependTestCase{
			{
				testCase: testCase{
					Id:             "prepend data to unknown key content, return not stored",
					Command:        "prepend test99 0 1 4\r\n9876\r\n",
					ExpectedResult: "NOT_STORED\r\n",
				},
				ExpectedKeyVal: "",
			},
			{
				testCase: testCase{
					Id:             "prepend data to existing key content, return stored",
					Command:        "prepend test4 0 1 4\r\nmore\r\n",
					ExpectedResult: "STORED\r\n",
				},
				ExpectedKeyVal: "more09182736",
			},
			{
				testCase: testCase{
					Id:             "prepend data to existing key content noreply, return nothing",
					Command:        "prepend test4 0 1 4 noreply\r\nmore\r\n",
					ExpectedResult: "",
				},
				ExpectedKeyVal: "moremore09182736",
			},
		}

		for _, tc := range prependCmdTestCases {
			t.Run(tc.Id, func(t *testing.T) {
				receivedData := dialConnectionAndCheckReceivedData(t, serverAddress, tc.testCase)

				if receivedData != tc.ExpectedResult {
					t.Errorf("expected to receive %s, got %s", tc.ExpectedResult, receivedData)
				}

				key := strings.Split(tc.Command, " ")[1]

				if tc.ExpectedKeyVal != "" {
					assertKeyContent(t, stubStore, key, tc.ExpectedKeyVal)
				}

			})
		}
	})
}
