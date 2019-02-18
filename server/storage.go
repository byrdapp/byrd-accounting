package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	s3Region = "eu-north-1"
	s3Bucket = "byrd-accounting-bucket"
)

// NewUpload -
func NewUpload(fileDir string) error {
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3Region),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		return err
	}
	sess := session.Must(s, err)
	if err := uploader(sess, fileDir); err != nil {
		return err
	}
	return nil
}

// Uploader S3 uploader
func uploader(s *session.Session, fileDir string) error {
	uploader := s3manager.NewUploader(s)
	file, err := os.Open(fileDir)
	if err != nil {
		return fmt.Errorf("Failed to open file: %q, %v", fileDir, err)
	}
	defer file.Close()

	stats, _ := file.Stat()
	size := stats.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:                 file,
		Bucket:               aws.String(s3Bucket),
		Key:                  aws.String(fileDir),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		return fmt.Errorf("Failed to upload file:  %v", err)
	}
	fmt.Printf("Successfully uploaded file to: %s\n", aws.StringValue(&result.Location))
	return nil
}
