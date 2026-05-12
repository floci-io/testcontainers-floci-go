package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

type Response struct {
	Message string `json:"message"`
}

func handler(_ context.Context) (Response, error) {
	return Response{Message: "hello from lambda"}, nil
}

func main() {
	lambda.Start(handler)
}
