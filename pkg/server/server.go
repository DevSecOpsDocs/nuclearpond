package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

type Request struct {
	Targets []string `json:"Targets"`
	Batches int      `json:"Batches"`
	Threads int      `json:"Threads"`
	Args    string   `json:"Args"`
	Output  string   `json:"Output"`
}

// Index
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// check if the key matches the generated API key
	if !checkAPIKey(r) {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, "Welcome to the API")
}

// Health check
func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func scanStatusHandler(w http.ResponseWriter, r *http.Request) {
	// check if the key matches the generated API key
	if !checkAPIKey(r) {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	scanId := parts[2]
	log.Println("Getting scan state for", scanId)

	state, err := getScanState(scanId)
	if err != nil {
		http.Error(w, "Error getting scan state: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": state})
}

// Scan
func scanHandler(w http.ResponseWriter, r *http.Request) {
	// check if the key matches the generated API key
	if !checkAPIKey(r) {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Received request", req)

	scanId := uuid.New().String()
	go backgroundScan(req, scanId)
	json.NewEncoder(w).Encode(map[string]string{"RequestId": scanId})
}

func serverCheck() {
	// check if NUCLEARPOND_API_KEY is set
	if os.Getenv("NUCLEARPOND_API_KEY") == "" {
		log.Println("NUCLEARPOND_API_KEY not set")
	}

	// check if AWS_REGION and AWS_LAMBDA_FUNCTION_NAME are set
	_, region := os.LookupEnv("AWS_REGION")
	if !region {
		log.Fatal("AWS_REGION not set")
		os.Exit(1)
	}
	_, function := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME")
	if !function {
		log.Fatal("AWS_LAMBDA_FUNCTION_NAME not set")
		os.Exit(1)
	}
	_, dynamodb := os.LookupEnv("AWS_DYNAMODB_TABLE")
	if !dynamodb {
		log.Fatal("AWS_DYNAMODB_TABLE not set")
		os.Exit(1)
	}
}

func HandleRequests() {
	// Check if the server is configured correctly
	serverCheck()

	// Retrieve/Generate API key
	generateAPIKey()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/health-check", healthHandler)
	http.HandleFunc("/scan", scanHandler)
	http.HandleFunc("/scan/", scanStatusHandler)

	http.ListenAndServe(":8080", nil)
}
