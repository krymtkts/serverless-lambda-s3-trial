package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

type OutEvent struct {
	Payload string `json:"payload"`
	Status  int    `json:"status"`
}

func readGzipedContent(body io.ReadCloser) (*bytes.Buffer, error) {
	rc := body
	gr, err := gzip.NewReader(rc)
	defer rc.Close()
	defer gr.Close()
	if err != nil {
		log.Fatal(err)
		return nil, errors.New("cannot read gziped content")
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(gr)
	return buf, nil
}

func makeGzipedContent(results []JSONAttributes) (*bytes.Buffer, error) {
	rtn := new(bytes.Buffer)
	gw := gzip.NewWriter(rtn)
	defer gw.Close()

	_, err := gw.Write([]byte("["))
	log.Print("write '['")
	if err != nil {
		log.Fatal(err)
		log.Fatal(results)
		return nil, errors.New("failed to write '['")
	}

	for _, result := range results {
		buf, err := json.Marshal(result)
		if err != nil {
			log.Fatal(err)
			log.Fatal(result)
			return nil, errors.New("failed to marshal result")
		}
		_, err = gw.Write(buf)
		if err != nil {
			log.Fatal(err)
			return nil, errors.New("failed to write buf")
		}
		_, err = gw.Write([]byte(","))
		if err != nil {
			log.Fatal(err)
			return nil, errors.New("failed to write ','")
		}
	}
	log.Print("marshal done")

	_, err = gw.Write([]byte("]"))
	log.Print("write ']'")
	if err != nil {
		log.Fatal(err)
		return nil, errors.New("failed to write ]")
	}

	return rtn, nil
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (OutEvent, error) {
	log.Print("read.")

	log.Print("read bucket objects.")
	inputBucketName := os.Getenv("InputBucketName")
	prefix := "download"
	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String(endpoints.ApNortheast1RegionID),
	})
	list, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &inputBucketName,
		Prefix: &prefix,
	})
	if err != nil {
		log.Fatal(err)
		return OutEvent{
			Payload: "Read & Write function is failed.",
			Status:  400,
		}, nil
	}
	var results []JSONAttributes
	for _, item := range list.Contents {
		obj, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: &inputBucketName,
			Key:    item.Key,
		})
		if err != nil {
			log.Fatal(err)
			return OutEvent{
				Payload: "Read & Write function is failed.",
				Status:  400,
			}, nil
		}

		buf, err := readGzipedContent(obj.Body)
		if err != nil {
			log.Fatal(err)
			return OutEvent{
				Payload: "Read & Write function is failed.",
				Status:  400,
			}, nil
		}

		var jsons []JSONAttributes
		err = json.Unmarshal(buf.Bytes(), &jsons)
		if err != nil {
			log.Fatal(err)
			return OutEvent{
				Payload: "Read & Write function is failed.",
				Status:  400,
			}, nil
		}
		results = append(results, jsons...)
	}

	log.Print("marshall json.")
	log.Print(len(results))

	buf, nil := makeGzipedContent(results)
	if err != nil {
		log.Fatal(err)
		return OutEvent{
			Payload: "Read & Write function is failed.",
			Status:  400,
		}, nil
	}

	log.Print("write to output bucket.")
	outputBucketName := os.Getenv("OutputBucketName")
	key := "go.json.gz"
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: &outputBucketName,
		Key:    &key,
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		log.Fatal(err)
		return OutEvent{
			Payload: "Read & Write function is failed.",
			Status:  400,
		}, nil
	}

	log.Print("end read.")
	return OutEvent{
		Payload: fmt.Sprintf("%d keys Read & Write function is finished.", list.KeyCount),
		Status:  200,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
