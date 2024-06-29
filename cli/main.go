package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"query/rag"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

func main() {
	// Parse command line arguments
	questionPtr := flag.String("question", "", "The question to ask the Lambda function")
	verbose := flag.Bool("verbose", false, "Show documents also")
	flag.Parse()

	if *questionPtr == "" {
		log.Fatalf("question parameter is required")
	}

	// Load the AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create a Lambda client
	client := lambda.NewFromConfig(cfg)

	// Define the payload
	payload := map[string]string{
		"question": *questionPtr,
	}

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("failed to marshal payload, %v", err)
	}

	// Create the Invoke input
	input := &lambda.InvokeInput{
		FunctionName: aws.String("hugoembedding"),
		Payload:      payloadBytes,
	}

	// Invoke the Lambda function
	result, err := client.Invoke(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to invoke lambda function, %v", err)
	}

	// Check for function error
	if result.FunctionError != nil {
		log.Fatalf("lambda function returned an error: %s", aws.ToString(result.FunctionError))
	}

	// Print the result
	var response rag.Response
	err = json.Unmarshal(result.Payload, &response)
	if err != nil {
		log.Fatalf("failed to unmarshal response payload, %v", err)
	}

	fmt.Println("Answer:", response.Answer)

	if *verbose {
		fmt.Println("\n The following documents were used \n ============\n")

		for _, doc := range response.Documents {
			fmt.Printf("Document ID: %d\n", doc.Id)
			fmt.Printf("Content: %s\n", doc.Content)
			fmt.Printf("Context: %s\n", doc.Context)
		}
	}
}
