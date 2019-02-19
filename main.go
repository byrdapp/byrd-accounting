package main

import (
	"log"

	"github.com/byblix/byrd-accounting/invoices"
	"github.com/joho/godotenv"
)

func main() {

	/**
	 * Run shellscript: $ sh run.sh for docker deploy
	 */

	if err := environment(); err != nil {
		log.Fatalf("Error with env: %s", err)
	}

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
