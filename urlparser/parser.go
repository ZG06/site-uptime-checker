package urlparser

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func ReadFile(filename string) ([]byte, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func ParseJSON(js []byte) ([]string, error) {
	var s []string
	if err := json.Unmarshal(js, &s); err != nil {
		return nil, err
	}

	return s, nil

}

// Read json file from the provided AWS S3 bucket and return it
func FetchS3Data(ctx context.Context, bucket, key string) ([]byte, error) {
	// Load AWS region and credentials.
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-central-1"),
	)
	if err != nil {
		return nil, err
	}

	// Create an S3 client.
	client := s3.NewFromConfig(cfg)

	// Get the json file with URLs from the S3 bucket.
	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		return nil, err
	}

	// Close response body after exiting the function.
	defer result.Body.Close()

	// Read the bytes from the json file and return the URLs.
	return io.ReadAll(result.Body)
}
