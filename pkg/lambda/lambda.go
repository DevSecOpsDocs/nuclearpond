package lambda

import (
        "encoding/json"
        "fmt"
        "log"
        "os"

        "github.com/DevSecOpsDocs/nuclearpond/pkg/outputs"
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
func InvokeLambdas(payload LambdaInvoke, lambda string, output string, region string) {
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
        response, err := invokeFunction(string(lambdaInvokeJson), lambda, region)
        if err != nil {
                fmt.Println(err)
        }

        // Parse lambda response Output
        var responseInterface interface{}
        json.Unmarshal([]byte(response), &responseInterface)
        // print responseInterface output value
        lambdaResponse := responseInterface.(map[string]interface{})["output"]

        // Change outputs depending on the output
        switch output {
        case "s3":
                outputs.S3Output(lambdaResponse)
        case "cmd":
                outputs.CmdOutput(lambdaResponse)
        case "json":
                outputs.JsonOutput(lambdaResponse)
        }
}

// Execute a lambda function and return the response
func invokeFunction(payload string, functionName string, region string) (string, error) {
        // Create a new session
        sess, err := session.NewSession(&aws.Config{
                Region: aws.String(region)}, // Using the passed region here
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
