package server

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Uploader S3 uploader
func Uploader(fileName string) error {
	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("Failed to open file: %q, %v", fileName, err)
	}
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("region=us-east-1")),
		Key:    aws.String("Test"),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("Failed to upload file:  %v", err)
	}
	fmt.Printf("Successfully uploaded file to: %s\n", aws.StringValue(&result.Location))
	return nil
}
