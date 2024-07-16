package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

type Input struct {
	Message string `json:"Message"`
}

type Output struct {
	Message string `json:"Message"`
}

func HandleRequest(ctx context.Context, input Input) (interface{}, error) {
	return Output{Message: input.Message}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
