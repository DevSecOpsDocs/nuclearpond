package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/DevSecOpsDocs/nuclearpond/pkg/core"
	"github.com/DevSecOpsDocs/nuclearpond/pkg/helpers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func backgroundScan(scanInput Request, scanId string) {
	targets := helpers.RemoveEmpty(scanInput.Targets)
	batches := helpers.SplitSlice(targets, scanInput.Batches)
	output := scanInput.Output
	threads := scanInput.Threads
	NucleiArgs := base64.StdEncoding.EncodeToString([]byte(scanInput.Args))
	silent := true

	functionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	regionName := os.Getenv("AWS_REGION")
	dynamodbTable := os.Getenv("AWS_DYNAMODB_TABLE")

	if functionName == "" || regionName == "" || dynamodbTable == "" {
		log.Fatal("Environment variables (AWS_LAMBDA_FUNCTION_NAME, AWS_REGION, AWS_DYNAMODB_TABLE) are not set.")
	}

	requestId := strings.ReplaceAll(scanId, "-", "")

	log.Println("Initiating scan with the id of ", scanId, "with", len(targets), "targets")
	storeScanState(requestId, "running")
	core.ExecuteScans(batches, output, functionName, NucleiArgs, threads, silent, regionName)
	storeScanState(requestId, "completed")
	log.Println("Scan", scanId, "completed")
}

func storeScanState(requestId string, status string) error {
	log.Println("Stored scan state in Dynamodb", requestId, "as", status)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return err
	}
	// Create DynamoDB client
	svc := dynamodb.New(sess)
	// Prepare the item to be put into the DynamoDB table
	item := &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("AWS_DYNAMODB_TABLE")),
		Item: map[string]*dynamodb.AttributeValue{
			"scan_id": {
				S: aws.String(requestId),
			},
			"status": {
				S: aws.String(status),
			},
			"timestamp": {
				N: aws.String(fmt.Sprintf("%d", time.Now().Unix())),
			},
			"ttl": {
				N: aws.String(fmt.Sprintf("%d", time.Now().Add(time.Duration(30*time.Minute)).Unix())),
			},
		},
	}
	// Store the item in DynamoDB
	_, err = svc.PutItem(item)
	if err != nil {
		log.Println("Failed to store scan state in Dynamodb:", err)
		return err
	}

	return nil
}

// function to retrieve the scan state from DynamoDB
func getScanState(requestId string) (string, error) {
	log.Println("Retrieving scan state from Dynamodb", requestId)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return "failed", err
	}
	// Create DynamoDB client
	svc := dynamodb.New(sess)
	// Prepare the item to be put into the DynamoDB table
	item := &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("AWS_DYNAMODB_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"scan_id": {
				S: aws.String(requestId),
			},
		},
	}
	// Store the item in DynamoDB
	result, err := svc.GetItem(item)
	if err != nil {
		return "failed", err
	}
	return *result.Item["status"].S, nil
}
