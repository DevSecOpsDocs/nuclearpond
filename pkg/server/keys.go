package server

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
)

func generateAPIKey() string {
	// if NUCLEARPOND_API_KEY is set, use that
	if os.Getenv("NUCLEARPOND_API_KEY") != "" {
		return os.Getenv("NUCLEARPOND_API_KEY")
	}
	// otherwise, generate a random API key
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	apiKey := hex.EncodeToString(b)
	log.Println("Generated API key:", apiKey)
	os.Setenv("NUCLEARPOND_API_KEY", apiKey)
	return apiKey
}

func checkAPIKey(r *http.Request) bool {
	apiKey := generateAPIKey()
	key := r.Header.Get("X-NuclearPond-API-Key")
	if key != apiKey {
		return false
	}
	return true
}
