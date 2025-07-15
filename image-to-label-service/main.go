package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// RPCRequest represents a JSON-RPC 2.0 request payload.
type RPCRequest struct {
	JsonRPC string      `json:"jsonrpc"` // Should always be "2.0"
	Method  string      `json:"method"`  // Method name, e.g., "classify"
	Params  interface{} `json:"params"`  // Arbitrary parameter object
	ID      int         `json:"id"`      // Unique request ID
}

// RPCResponse represents a JSON-RPC 2.0 response.
type RPCResponse struct {
	JsonRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
	ID      int             `json:"id"`
}

// dialWithRetry attempts to open a TCP connection with exponential backoff.
// It retries up to `maxAttempts` times with an initial delay of `initialDelay`.
func dialWithRetry(address string, maxAttempts int, initialDelay time.Duration) (net.Conn, error) {
	var conn net.Conn
	var err error
	delay := initialDelay
	for i := 1; i <= maxAttempts; i++ {
		conn, err = net.Dial("tcp", address)
		if err == nil {
			return conn, nil
		}
		log.Printf("Attempt %d: waiting %v, error: %v\n", i, delay, err)
		time.Sleep(delay)
		delay *= 2
	}
	return nil, fmt.Errorf("could not connect after %d attempts: %w", maxAttempts, err)
}

// getLabelSet returns the fixed classification label sets for type, era, and occasion.
func getLabelSet() map[string][]string {
	return map[string][]string{
		"type": {
			"t-shirt", "long sleeve top", "tank top", "crop top",
			"shirt", "blouse", "sweater", "hoodie", "cardigan",
			"blazer", "trench coat", "denim jacket", "leather jacket", "puffer jacket",
			"mini skirt", "midi skirt", "maxi skirt",
			"jeans", "cargo pants", "tailored trousers",
			"maxi dress", "midi dress", "sundress",
			"jumpsuit", "romper",
			"bra", "bralette", "thong", "briefs", "bikini bottom",
		},
		"era": {
			"minimalist", "grunge", "y2k", "romantic", "boho",
			"90s", "80s", "70s", "clean girl", "old money",
			"streetwear", "dark academia", "mod", "preppy",
		},
		"occasion": {
			"casual", "work", "party", "date night",
			"beach", "vacation", "lounge", "festival", "formal",
		},
	}
}

// classifyHandler handles POST /classify and relays the image to the CLIP service via JSON-RPC.
// It responds with the classification result or an appropriate error.
func classifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get image file
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read image bytes
	imgBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Read error", http.StatusInternalServerError)
		return
	}

	// Encode to base64 for JSON-RPC
	imgBase64 := base64.StdEncoding.EncodeToString(imgBytes)
	params := map[string]any{
		"image":  imgBase64,
		"labels": getLabelSet(),
	}

	// Determine backend CLIP service address
	clipAddress := os.Getenv("CLIP_SERVICE_ADDRESS")
	if clipAddress == "" {
		clipAddress = "clip-service:9000"
	}

	// Retry dial
	conn, err := dialWithRetry(clipAddress, 6, 1*time.Second)
	if err != nil {
		http.Error(w, "clip-service not available", http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	// Prepare JSON-RPC request
	req := RPCRequest{
		JsonRPC: "2.0",
		Method:  "classify",
		Params:  params,
		ID:      1,
	}

	data, err := json.Marshal(req)
	if err != nil {
		http.Error(w, "Failed to encode request", http.StatusInternalServerError)
		return
	}

	// Send to backend
	if _, err := conn.Write(append(data, '\n')); err != nil {
		http.Error(w, "Failed to write to backend", http.StatusInternalServerError)
		return
	}

	// Read response from backend
	buf := make([]byte, 4096*8)
	n, err := conn.Read(buf)
	if err != nil {
		http.Error(w, "Failed to read from backend", http.StatusInternalServerError)
		return
	}

	// Return backend response
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf[:n])
}

// serveIndex serves the HTML UI from the public folder.
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./public/index.html")
}

// main sets up the HTTP server for the fashion app.
func main() {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/classify", classifyHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Fashion app listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
