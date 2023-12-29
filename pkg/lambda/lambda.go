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

// LambdaInvoke estrutura para invocar a função lambda.
type LambdaInvoke struct {
    Targets []string `json:"Targets"`
    Args    []string `json:"Args"`
    Output  string   `json:"Output"`
}

// InvokeLambdas prepara e executa a função lambda.
func InvokeLambdas(payload LambdaInvoke, lambdaFunction string, output string, region string) {
    if len(payload.Targets) == 0 {
        return
    }

    lambdaInvokeJson, err := json.Marshal(payload)
    if err != nil {
        log.Fatal(err)
    }

    response, err := invokeFunction(string(lambdaInvokeJson), lambdaFunction, region)
    if err != nil {
        log.Println("Erro ao invocar a função lambda:", err)
        return
    }

    var responseInterface map[string]interface{}
    err = json.Unmarshal([]byte(response), &responseInterface)
    if err != nil {
        log.Println("Erro ao deserializar a resposta da função lambda:", err)
        return
    }

    lambdaResponse, exists := responseInterface["output"]
    if !exists {
        log.Println("Chave 'output' não encontrada na resposta da função lambda")
        return
    }

    switch output {
    case "s3":
        outputs.S3Output(lambdaResponse)
    case "cmd":
        outputs.CmdOutput(lambdaResponse)
    case "json":
        outputs.JsonOutput(lambdaResponse)
    }
}

// invokeFunction executa uma função lambda e retorna a resposta.
func invokeFunction(payload string, functionName string, region string) (string, error) {
    sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
    if err != nil {
        return "", fmt.Errorf("Erro ao criar sessão AWS: %v", err)
    }

    svc := lambda.New(sess)

    input := &lambda.InvokeInput{
        FunctionName: aws.String(functionName),
        Payload:      []byte(payload),
    }

    result, err := svc.Invoke(input)
    if err != nil {
        return "", fmt.Errorf("Erro ao invocar função lambda: %v", err)
    }

    return string(result.Payload), nil
}
