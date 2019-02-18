package main

import (
	"log"

	"github.com/byblix/byrd-accounting/invoices"
	"github.com/joho/godotenv"
)

func main() {

	/**
	 * Run shellscript: $ sh run.sh
	 */

	if err := environment(); err != nil {
		log.Fatalf("Error with env: %s", err)
	}
	// invoices.ExamplePdf()
	invoices.InitInvoiceOutput()
	// server.Start()
	// if err := server.Uploader("pdf.pdf"); err != nil {
	// 	log.Fatalln(err)
	// }
}

func environment() error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	return nil
}
