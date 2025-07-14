package main

import (
	"encoding/json"
	"fmt"
	"net"
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

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
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
