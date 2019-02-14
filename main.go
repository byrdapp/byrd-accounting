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
	Pagination *Pagination      `json:"pagination,omitempty"`
}

// Pagination for getting more invoices
type Pagination struct {
	PageSize  int    `json:"pageSize,omitempty"`
	Results   int    `json:"results,omitempty"` //Total results
	FirstPage string `json:"firstPage,omitempty"`
	NextPage  string `json:"nextPage,omitempty"`
	LastPage  string `json:"lastPage,omitempty"`
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
	LineNumber  byte    `json:"lineNumber,omitempty"`  /*MUST be #2 on voice*/
	Description string  `json:"description,omitempty"` /*If this == dragonplan*/
	Quantity    float32 `json:"quantity,omitempty"`    /*Number of credits*/
}

/**
 * 1. Get all invoiceNumbers => GET: /invoices/booked
 * 2. Get all invoices by number => GET invoices/booked/{num}
 * 3. Store Lines info ()
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
	invoice, err := getEconomicsBookedInvoice(invoices)
	if err != nil {
		log.Fatalf("Couldnt get the booked invoices struct: %s", err)
	}
	_ = invoice
	// fmt.Printf("%+v \n", invoice)

}

func getEconomicsBookedInvoices() (*BookedInvoices, error) {
	invoices := BookedInvoices{}
	url := ecoURL + "/invoices/booked"
	res := createReq(url)
	err := json.NewDecoder(res.Body).Decode(&invoices)
	if err != nil {
		return nil, err
	}
	return &invoices, nil
}

func getEconomicsBookedInvoice(invoices *BookedInvoices) (*BookedInvoice, error) {
	invoice := BookedInvoice{}
	for idx, val := range invoices.Collection {
		fmt.Printf("Getting #%s with invoiceNum: %s", strconv.Itoa(idx), strconv.Itoa(val.BookedInvoiceNumber))
		url := ecoURL + "/invoices/booked/" + strconv.Itoa(val.BookedInvoiceNumber)
		res := createReq(url)
		err := json.NewDecoder(res.Body).Decode(&invoice)
		if err != nil {
			return nil, err
		}
		printStructAsJSONText(invoice.Lines)
	}
	return &invoice, nil
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
