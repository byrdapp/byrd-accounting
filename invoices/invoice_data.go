package invoices

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/byblix/byrd-accounting/server"
)

// BookedInvoices - endpoint https://restapi.e-conomic.com/invoices/booked
type BookedInvoices struct {
	Collection []*BookedInvoiceNumber `json:"collection,omitempty"`
	Pagination *Pagination            `json:"pagination,omitempty"`
}

// BookedInvoiceNumber -
type BookedInvoiceNumber struct {
	BookedInvoiceNumber int `json:"bookedInvoiceNumber,omitempty"`
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
	BookedInvoiceNumber int        `json:"bookedInvoiceNumber,omitempty"`
	Date                string     `json:"date,omitempty"`
	Currency            string     `json:"currency,omitempty"`
	NetAmount           float64    `json:"netAmount,omitempty"`
	GrossAmount         float64    `json:"grossAmount,omitempty"`
	VatAmount           float64    `json:"vatAmount,omitempty"`
	Lines               []*Lines   `json:"lines,omitempty"`
	Recipient           *Recipient `json:"recipient"`
}

// Recipient -
type Recipient struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Country string `json:"country"`
}

// Lines -
type Lines struct {
	LineNumber     byte    `json:"lineNumber,omitempty"`  /*MUST be #2 on voice*/
	Description    string  `json:"description,omitempty"` /*If this == dragonplan*/
	CreditQuantity float64 `json:"quantity,omitempty"`    /*Number of credits*/
}

// DateRange -
type DateRange struct {
	From  string
	To    string
	Query string
}

var (
	logger *log.Logger
	ecoURL = "https://restapi.e-conomic.com"
)

const (
	creditLineNumber = 2
	photographerCut  = 15
)

// InitInvoiceOutput starts the whole thing :-)
func InitInvoiceOutput() error {
	// testinterval := "date$gte:2019-02-13$and:date$lte:2019-02-28"
	d := &DateRange{}
	interval := d.setDateRange()
	invoices, err := getEconomicsBookedInvoices(interval)
	if err != nil {
		log.Fatalf("Couldnt get the booked invoices: %s", err)
		return err
	}

	specificInvoices, err := getSpecificEcoBookedInvoices(invoices.Collection)
	if err != nil {
		log.Fatalf("Error with getting specific invoice: %s", err)
	}
	// For each invoices (BookedINvoices), fetch the corresponding specific invoice line /invoices/booked/{number}
	fileName, err := createPDFFromInvoice(specificInvoices)
	if err != nil {
		log.Fatalf("Couldnt create PDF with http: %s", err)
		return err
	}
	if err := server.NewUpload(fileName); err != nil {
		log.Fatalf("couldt upload to server: %s", err)
		return err
	}
	return nil
}

func getEconomicsBookedInvoices(query string) (*BookedInvoices, error) {
	//syntax: https://restdocs.e-conomic.com/#filter-operators
	// combined mongo query ex.: date$gte:2018-01-01$and:date$lte:2018-01-09
	invoices := BookedInvoices{}
	url := ecoURL + "/invoices/booked"
	res := createReq(url, query)
	err := json.NewDecoder(res.Body).Decode(&invoices)
	if err != nil {
		return nil, err
	}
	return &invoices, nil
}

func getSpecificEcoBookedInvoices(invoiceNums []*BookedInvoiceNumber) ([]*BookedInvoice, error) {
	specificInvoices := []*BookedInvoice{}
	for _, num := range invoiceNums {
		invoice := &BookedInvoice{}
		url := ecoURL + "/invoices/booked/" + strconv.Itoa(num.BookedInvoiceNumber)
		res := createReq(url, "")
		err := json.NewDecoder(res.Body).Decode(&invoice)
		if err != nil {
			return nil, err
		}
		fmt.Println(invoice.BookedInvoiceNumber)
		// Set PDFData values based on invoice values
		specificInvoices = append(specificInvoices, invoice)
	}
	return specificInvoices, nil
}

func createPDFFromInvoice(invoices []*BookedInvoice) (string, error) {
	fileName, err := WriteInvoicesPDF(invoices, getMonthAgo())
	if err != nil {
		return "", err
	}
	return fileName, nil
}

func createReq(url string, params string) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	header := &req.Header
	query := req.URL.Query()
	if err != nil {
		log.Fatalf("Error with request setup: %s", err)
	}
	header.Add("Content-Type", "application/json")
	header.Add("X-AppSecretToken", os.Getenv("ECONOMIC_SECRET_TOKEN"))
	header.Add("X-AgreementGrantToken", os.Getenv("ECONOMIC_PUBLIC_TOKEN"))
	query.Add("sort", "date")
	if params != "" {
		query.Add("filter", params)
	}
	req.URL.RawQuery = query.Encode()
	res, err := client.Do(req)
	fmt.Println(req.URL.String())
	if err != nil {
		log.Fatalf("Error with client HTTP: %s", err)
	}
	// defer res.Body.Close()
	return res
}

func getDayAgo() string {
	t := time.Now().UTC()
	hour := time.Hour
	t.Add(-hour)
	t.Format("20060102")
	return t.String()[:10]
}

func getMonthAgo() string {
	t := time.Now().UTC()
	month := t.AddDate(0, -1, 0)
	month.Format("20060102")
	return month.String()[:10]
}

func (d *DateRange) setDateRange() string {
	dates := DateRange{
		From: getMonthAgo(),
		To:   getDayAgo(),
	}
	dates.Query = "date$gte:" + dates.From + "$and:date$lte:" + dates.To
	fmt.Printf("Interval from: %s\n to: %s\n", dates.From, dates.To)
	return dates.Query
}

func printStructAsJSONText(i interface{}) {
	format, _ := json.Marshal(i)
	fmt.Println(string(format))
}
