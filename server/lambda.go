package server

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

/**
 * cron: 0 0 12 LW * ? (Every month on the last weekday, at noon)
 * GOOS=linux GOARCH=amd64 go build -o main main.go
 * zip main.zip main
 */

// Request -
type Request struct {
	ID    float64 `json:"id"`
	Value string  `json:"value"`
}

// Response -
type Response struct {
	Message string `json:"message"`
	Ok      bool   `json:"ok"`
}

// Start the handler for storing invoice data
func Start() {
	lambda.Start(HandleRequest)
}
// HandleRequest -
func HandleRequest(ctx context.Context) (*Response, error) {
	DoEveryThing()
	return res, nil
}

/**
 * 1. Get all invoices from for the previous month
 * 2. Convert to PDF
 * 3. Upload to S3
 * 4. Log("Success") or error
 */
