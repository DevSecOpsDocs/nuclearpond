package lambda

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// struct for lambda invoke
type LambdaInvoke struct {
	Targets []string `json:"Targets"`
	Args    []string `json:"Args"`
	Output  string   `json:"Output"`
}

// Stage the lambda function for executing
func InvokeLambdas(payload LambdaInvoke, lambda string, output string) {
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
			log.Println("Scan completed with no findings")
			return
		}
		log.Println("Saved results in", lambdaResponse)
	} else if output == "cmd" {
		// convert lambdaResponse to string
		lambdaResponseString := lambdaResponse.(string)
		// convert response from base64 to colorized terminal output
		decodedBytes, _ := base64.StdEncoding.DecodeString(lambdaResponseString)
		// if decodedBytes is empty then return
		if len(decodedBytes) == 0 {
			log.Println("Scan completed with no output")
			return
		}
		log.Println("Scan complete with output:")
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

// Execute a lambda function and return the response
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
		log.Fatal("Failed to invoke lambda function: ", err)
		os.Exit(1)
	}

	// Return the response
	return string(result.Payload), nil
}
