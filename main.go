package main

import (
	"log"

	"github.com/byblix/byrd-accounting/invoices"
	"github.com/joho/godotenv"
)

/**
 * cron: 0 0 12 LW * ? (Every month on the last weekday, at noon)
 * GOOS=linux GOARCH=amd64 go build -o main main.go
 * zip main.zip main
 */

func main() {
	/* Run shellscript: `$ sh create-lambda.sh` for docker deploy */
	if err := environment(); err != nil {
		log.Printf("Error with env: %s", err)
	}
	if err := invoices.InitInvoiceOutput(); err != nil {
		log.Fatalf("Error occurred: %s", err)
	}
}

// func main() {
// 	/* Run shellscript: `$ sh create-lambda.sh` for docker deploy */
// 	if err := environment(); err != nil {
// 		log.Printf("Error with env: %s", err)
// 	}
// 	lambda.Start(HandleRequest)
// }

// HandleRequest -
func HandleRequest() {
	log.Println("Starting PDF'er")
	if err := invoices.InitInvoiceOutput(); err != nil {
		log.Fatalf("Error occurred: %s", err)
	}
}

func environment() error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	return nil
}
