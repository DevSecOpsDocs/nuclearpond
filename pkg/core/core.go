package core

import (
	"encoding/base64"
	"log"
	"strings"
	"sync"

	"github.com/DevSecOpsDocs/nuclearpond/pkg/lambda"
)

func ExecuteScans(batches [][]string, output string, lambdaName string, nucleiArgs string, threads int, silent bool) {
	// Create a WaitGroup to wait for the goroutines to finish
	var wg sync.WaitGroup

	// Set the number of threads to use
	numThreads := threads

	// Create a channel to pass tasks to the goroutines
	tasks := make(chan func())

	// Start the goroutines
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func() {
			for task := range tasks {
				task()
			}
			wg.Done()
		}()
	}

	// Add tasks to the channel
	for _, batch := range batches {
		tasks <- func() {
			decodedBytes, _ := base64.StdEncoding.DecodeString(nucleiArgs)
			// Convert args.NucleiArg from base64, to string, split by space, and convert to list
			nucleiFlags := strings.Split(string(decodedBytes), " ")

			// create lambda invoke struct
			lambdaInvoke := lambda.LambdaInvoke{
				Targets: batch,
				Args:    nucleiFlags,
				Output:  output,
			}
			lambda.InvokeLambdas(lambdaInvoke, lambdaName, output)
		}
	}

	close(tasks)
	wg.Wait()

	// Print the results if not silent mode
	if !silent {
		log.Println("Completed all parallel operations, best of luck!")
	}
}
