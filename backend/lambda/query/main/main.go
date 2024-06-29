package main

import (
	"context"
	re "ragembeddings"

	"ragembeddings/query"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(Handler)
}
func Handler(ctx context.Context, event re.QueryRequest) (re.Response, error) {
	response := query.Query(ctx, event)
	return response, nil
}
