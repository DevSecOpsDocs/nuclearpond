package outputs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// function for s3 output
func S3Output(lambdaResponse interface{}) {
	if strings.Contains(lambdaResponse.(string), "No findings") {
		log.Println("Scan completed with no findings")
		return
	}
	log.Println("Saved results in", lambdaResponse)
}

// function for cmd output
func CmdOutput(lambdaResponse interface{}) {
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
}

// function for json output
func JsonOutput(lambdaResponse interface{}) {
	// if lambdaResponse contains the string "No findings" then return
	if strings.Contains(lambdaResponse.(string), "No findings") {
		log.Println("Scan completed with no findings")
		return
	}
	// pretty print json in lambdaResponse indented by 4 spaces
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, []byte(lambdaResponse.(string)), "", "    ")
	if error != nil {
		log.Println("JSON parse error: ", error)
		return
	}
	// if results.json exists, append to it
	if _, err := os.Stat("results.json"); err == nil {
		f, err := os.OpenFile("results.json", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		_, err = f.Write(prettyJSON.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Appended results in results.json")
		return
	}
	// write prettyJSON to file
	f, err := os.Create("results.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.Write(prettyJSON.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Saved results in results.json")
}
