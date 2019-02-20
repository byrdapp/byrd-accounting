package server

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	s3Region = "eu-north-1"
	s3Bucket = "byrd-accounting-bucket"
)

// NewUpload -
func NewUpload(file []byte, dateStamp string) error {
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3Region),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS"), os.Getenv("AWS_SECRET"), ""),
	})
	if err != nil {
		return err
	}
	sess := session.Must(s, err)
	if err := uploader(sess, file, dateStamp); err != nil {
		return err
	}
	return nil
}

// Uploader S3 uploader
func uploader(s *session.Session, file []byte, dateStamp string) error {
	uploader := s3manager.NewUploader(s)
	fileName := dateStamp[:7] + ".pdf"
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:   bytes.NewBuffer(file),
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(string(fileName)),
		// ContentType:          aws.String(http.DetectContentType(buffer)),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		return fmt.Errorf("Failed to upload file:  %v", err)
	}
	fmt.Printf("Successfully uploaded file to: %s\n", aws.StringValue(&result.Location))
	return nil
}
