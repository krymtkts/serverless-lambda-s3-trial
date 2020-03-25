package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// OutEvent defines your lambda output data structure,
type OutEvent struct {
	Payload string `json:"payload"`
	Status  int    `json:"status"`
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (OutEvent, error) {
	log.Print("entry point.")

	log.Print("cleanup output bucket.")
	outputBucketName := os.Getenv("BUCKET")
	KEY := "go.json.gz"
	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String(endpoints.ApNortheast1RegionID),
	})
	svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &outputBucketName,
		Key:    &KEY,
	})

	return OutEvent{
		Payload: "Entry function is finished.",
		Status:  200,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
