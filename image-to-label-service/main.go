package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type RPCRequest struct {
	JsonRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type RPCResponse struct {
	JsonRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
	ID      int             `json:"id"`
}

func dialWithRetry(address string, maxAttempts int, initialDelay time.Duration) (net.Conn, error) {
	var conn net.Conn
	var err error
	delay := initialDelay
	for i := 1; i <= maxAttempts; i++ {
		conn, err = net.Dial("tcp", address)
		if err == nil {
			return conn, nil
		}
		fmt.Printf("Attempt %d: waiting %v, error: %v\n", i, delay, err)
		time.Sleep(delay)
		delay *= 2 // exponential backoff
	}
	return nil, fmt.Errorf("could not connect after %d attempts: %w", maxAttempts, err)
}

func main() {
	conn, err := dialWithRetry("clip-service:9000", 6, 1*time.Second)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	req := RPCRequest{
		JsonRPC: "2.0",
		Method:  "add",
		Params:  map[string]int{"a": 5, "b": 7},
		ID:      1,
	}

	data, _ := json.Marshal(req)
	conn.Write(data)

	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	fmt.Println("Raw response:", string(buf[:n]))

	var resp RPCResponse
	if err := json.Unmarshal(buf[:n], &resp); err != nil {
		panic(err)
	}
	fmt.Printf("Result: %s\n", resp.Result)
}
