package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var (
	logger *log.Logger
	ecoURL = "https://restapi.e-conomic.com"
)

// BookedInvoices - endpoint https://restapi.e-conomic.com/invoices/booked
type BookedInvoices struct {
	Collection []*BookedInvoice `json:"collection,omitempty"`
}

// BookedInvoice - endpoint https://restapi.e-conomic.com/invoices/booked/:number
type BookedInvoice struct {
	BookedInvoiceNumber string   `json:"bookedInvoiceNumber,omitempty"`
	Date                string   `json:"date,omitempty"`
	Currency            string   `json:"currency,omitempty"`
	NetAmount           int      `json:"netAmount,omitempty"`
	GrossAmount         int      `json:"grossAmount,omitempty"`
	VatAmount           int      `json:"vatAmount,omitempty"`
	Lines               []*Lines `json:"lines,omitempty"`
}

// Lines -
type Lines struct {
	LineNumber  byte   `json:"lineNumber,omitempty"`  /*MUST be #2 on voice*/
	Description string `json:"description,omitempty"` /*If this == dragonplan*/
	Quantity    int    `json:"quantity,omitempty"`    /*Number of credits*/
}

/**
 * 1. Get all invoiceNumbers => GET: /invoices/booked
 * 2. If DragonPlan set it to unlimited
 */

func main() {
	if err := environment(); err != nil {
		logger.Fatalf("Error with env: %s", err)
	}

}

// func calculateTax(value float32) (float32, error) {
// 	if value < 0 {
// 		return 0, errors.New("Must be a pos int")
// 	}
// 	return value * 1.25, nil
// }

func getEconomicsBookedInvoices() (*BookedInvoices, error) {
	url := ecoURL + "/invoices/booked"
	invoices := BookedInvoices{}
	res := createReqGetDecoder(url)
	_ = res
	return &invoices, nil
}

func getEconomicsBookedInvoice() {
	invoice := BookedInvoice{}
	_ = invoice
}

func createReqGetDecoder(url string) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error with request setup: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-AppSecretToken", os.Getenv("X-AppSecretToken"))
	req.Header.Add("X-AgreementGrantToken", os.Getenv("ECONOMIC_PUBLIC_TOKEN"))
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error with client HTTP: %s", err)
	}
	defer res.Body.Close()
	return nil
}

func environment() error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	fmt.Println("Loaded the .env file")
	return nil
}
