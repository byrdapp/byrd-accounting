package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
	BookedInvoiceNumber int      `json:"bookedInvoiceNumber,omitempty"`
	Date                string   `json:"date,omitempty"`
	Currency            string   `json:"currency,omitempty"`
	NetAmount           float32  `json:"netAmount,omitempty"`
	GrossAmount         float32  `json:"grossAmount,omitempty"`
	VatAmount           float32  `json:"vatAmount,omitempty"`
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
	// Get economics invoices data requests => struct
	invoices, err := getEconomicsBookedInvoices()
	if err != nil {
		log.Fatalf("Couldnt get the booked invoices struct: %s", err)
	}

	// For each invoices (BookedINvoices), fetch the corresponding specific invoice line /invoices/booked/{number}
	getEconomicsBookedInvoice(invoices)
}

func getEconomicsBookedInvoices() (*BookedInvoices, error) {
	invoices := BookedInvoices{}
	url := ecoURL + "/invoices/booked"
	res := createReq(url)
	err := json.NewDecoder(res.Body).Decode(&invoices)
	if err != nil {
		return nil, err
	}
	printStructAsJSONText(invoices)
	return &invoices, nil
}

func getEconomicsBookedInvoice(invoices *BookedInvoices) error {
	invoice := BookedInvoice{}
	_ = invoice
	for idx, val := range invoices.Collection {
		fmt.Println("Getting: ", idx, val.BookedInvoiceNumber)
		url := ecoURL + "invoices/booked" + strconv.Itoa(val.BookedInvoiceNumber)
		res := createReq(url)
		err := json.NewDecoder(res.Body).Decode(&invoice)
		if err != nil {
			return err
		}
		printStructAsJSONText(invoices)
	}

	return nil
}

func createReq(url string) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error with request setup: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-AppSecretToken", os.Getenv("ECONOMIC_SECRET_TOKEN"))
	req.Header.Add("X-AgreementGrantToken", os.Getenv("ECONOMIC_PUBLIC_TOKEN"))
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error with client HTTP: %s", err)
	}
	// defer res.Body.Close()
	return res
}

func environment() error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	return nil
}

func printStructAsJSONText(i interface{}) {
	format, _ := json.Marshal(i)
	fmt.Println(string(format))
}
