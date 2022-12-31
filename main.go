package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/common-nighthawk/go-figure"
)

var args struct {
	Targets   string `arg:"-l,--list" help:"input file with list of urls"`
	Target    string `arg:"-t,--target" help:"target url for individual targets"`
	NucleiArg string `arg:"-a,separate" help:"Nuclei arguments base64 encoded to be sent`
	Region    string `arg:"-r,--region" help:"aws region"`
	Lambda    string `arg:"-f,--lambda-function" help:"lambda function name"`
	Silent    string `arg:"-s,--silent" help:"silent mode true/false"`
	BatchSize int    `arg:"-b,--batchsize" help:"batch size"`
	Output    string `arg:"-o,--output" help:"output type (s3, stdout, or json)"`
	Threads   int    `arg:"-n,--threads" help:"number of threads to use"`
}

// struct for lambda invoke
type lambdaInvoke struct {
	Targets []string `json:"Targets"`
	Args    []string `json:"Args"`
	Output  string   `json:"Output"`
}

func main() {
	arg.MustParse(&args)

	// Decode nuclei args
	decodedArgs, _ := base64.StdEncoding.DecodeString(args.NucleiArg)

	// if not silent mode, print banner
	if args.Silent == "" {
		// Print ascii art
		myFigure := figure.NewFigure("Nuclear Pond", "", true)
		myFigure.Print()
		fmt.Println()
		fmt.Println("Nuclear Pond is a tool to run nuclei in parallel on AWS Lambda")
		fmt.Println("Version: 0.1")
		fmt.Println("Author: @jonathanwalker")
		fmt.Println()
		fmt.Println("Configuration")
		fmt.Println("Function: ", args.Lambda)
		fmt.Println("Region: ", args.Region)
		if args.Targets == "" {
			fmt.Println("Arguments: nuclei -t", args.Target, string(decodedArgs))
		} else {
			fmt.Println("Arguments: nuclei -l", args.Targets, string(decodedArgs))
		}

		fmt.Println("Batch Size: ", args.BatchSize)
		fmt.Println("Threads: ", args.Threads)
		fmt.Println("Output: ", args.Output)

		fmt.Println()
	}

	// Check if either Target or Targets, at least one is required
	if args.Target == "" && args.Targets == "" {
		fmt.Println("Please specify either a target url or a list of targets")
		os.Exit(1)
	}

	// if -l is specified, read the file and return a list of urls
	var urls []string
	if args.Targets != "" {
		urls = readUrls(args.Targets)
	} else {
		urls = append(urls, args.Target)
	}

	// Check if lambda function name is specified
	if args.Lambda == "" {
		fmt.Println("Please specify a the backend lambda function name")
		os.Exit(1)
	}

	// Check if batches is > 0, if not exit with error
	if args.BatchSize <= 0 {
		fmt.Println("Please specify a batch size greater than 0")
		os.Exit(1)
	}

	// If threads is not specified, set it to 5
	if args.Threads == 0 {
		args.Threads = 5
	}

	// remove empty strings item from slice urls
	urls = removeEmpty(urls)
	batches := splitSlice(urls, args.BatchSize)
	fmt.Println("Launching...")
	fmt.Println("Total targets: ", len(urls))
	fmt.Println("Number of Invocations: ", len(batches))
	fmt.Println()

	// Create a WaitGroup to wait for the goroutines to finish
	var wg sync.WaitGroup

	// Set the number of threads to use
	numThreads := 5

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
			decodedBytes, _ := base64.StdEncoding.DecodeString(args.NucleiArg)
			// Convert args.NucleiArg from base64, to string, split by space, and convert to list
			nucleiFlags := strings.Split(string(decodedBytes), " ")

			// create lambda invoke struct
			lambdaInvoke := lambdaInvoke{
				Targets: batch,
				Args:    nucleiFlags,
				Output:  args.Output,
			}
			invokeLambdas(lambdaInvoke, args.Lambda, args.Output)
		}
	}

	close(tasks)
	wg.Wait()

	// Print the results if not silent mode
	if args.Silent == "" {
		fmt.Println("Completed all parallel operations, best of luck!")
	}
}

func invokeLambdas(payload lambdaInvoke, lambda string, output string) {
	// Bug to fix another day
	if payload.Targets[0] == "" {
		return
	}

	// convert lambdaInvoke to json string
	lambdaInvokeJson, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	// invoke lambda function
	response, err := invokeFunction(string(lambdaInvokeJson), lambda)
	if err != nil {
		fmt.Println(err)
	}

	// Parse lambda response Output
	var responseInterface interface{}
	json.Unmarshal([]byte(response), &responseInterface)
	// print responseInterface output value
	lambdaResponse := responseInterface.(map[string]interface{})["output"]

	// Change outputs depending on the output type
	if output == "s3" {
		// if lambdaResponse contains the string "No findings" then return
		if strings.Contains(lambdaResponse.(string), "No findings") {
			return
		}
		fmt.Println("Results stored in:", lambdaResponse)
	} else if output == "cmd" {
		// convert lambdaResponse to string
		lambdaResponseString := lambdaResponse.(string)
		// convert response from base64 to colorized terminal output
		decodedBytes, _ := base64.StdEncoding.DecodeString(lambdaResponseString)
		fmt.Println(string(decodedBytes))
	} else if output == "json" {
		// if lambdaResponse contains the string "No findings" then return
		if strings.Contains(lambdaResponse.(string), "No findings") {
			return
		}
		// pretty print json in lambdaResponse indented by 4 spaces
		var prettyJSON bytes.Buffer
		error := json.Indent(&prettyJSON, []byte(lambdaResponse.(string)), "", "    ")
		if error != nil {
			log.Println("JSON parse error: ", error)
			return
		}
		fmt.Println(string(prettyJSON.Bytes()))
	}
}

func invokeFunction(payload string, functionName string) (string, error) {
	// Create a new session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create a Lambda service client.
	svc := lambda.New(sess)

	// Create the input
	input := &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      []byte(payload),
	}

	// Invoke the lambda function
	result, err := svc.Invoke(input)
	if err != nil {
		return "", err
	}

	// Return the response
	return string(result.Payload), nil
}

// function to read a file and return urls in a list
func readUrls(filename string) []string {
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

// Split a list into batches
func splitSlice(items []string, batchSize int) [][]string {
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

// Remove empty items
func removeEmpty(items []string) []string {
	var result []string

	for _, item := range items {
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}
