package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		delay *= 2
	}
	return nil, fmt.Errorf("could not connect after %d attempts: %w", maxAttempts, err)
}

func main() {
	// Read image file (use any .jpg/.png)
	imgBytes, err := ioutil.ReadFile("myimage.jpg")
	if err != nil {
		panic(err)
	}
	imgBase64 := base64.StdEncoding.EncodeToString(imgBytes)

	params := map[string]interface{}{
		"image": imgBase64,
		"labels": map[string][]string{
			"type":     {"mini skirt", "jeans", "tank top", "blazer", "maxi dress", "puffer jacket", "cardigan"},
			"era":      {"grunge", "y2k", "romantic", "boho", "90s", "mod", "minimalist"},
			"occasion": {"party", "casual", "work", "beach", "sports", "lounge", "festival"},
		},
	}

	conn, err := dialWithRetry("clip-service:9000", 6, 1*time.Second)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	req := RPCRequest{
		JsonRPC: "2.0",
		Method:  "classify",
		Params:  params,
		ID:      1,
	}

	data, _ := json.Marshal(req)
	conn.Write(append(data, '\n'))

	buf := make([]byte, 4096*8)
	n, _ := conn.Read(buf)
	fmt.Println("Raw response:", string(buf[:n]))

	var resp RPCResponse
	if err := json.Unmarshal(buf[:n], &resp); err != nil {
		panic(err)
	}
	fmt.Printf("Result: %s\n", resp.Result)
}
