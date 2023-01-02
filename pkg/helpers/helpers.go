package helpers

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// function to read a file and return urls in a list
func ReadUrlsFromFile(filename string) []string {
	// Open a txt file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read the file
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the data to a string
	text := string(data)

	// Split the string into a list of urls
	urls := strings.Split(text, "\n")

	return urls
}

// Splits targets into batches by returning a slice of slices
func SplitSlice(items []string, batchSize int) [][]string {
	var batches [][]string

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	return batches
}

// Remove empty items for a slice of strings
func RemoveEmpty(items []string) []string {
	var result []string

	for _, item := range items {
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}
