package main

import (
	"github.com/byblix/byrd-accounting/server"
)

func main() {

	/**
	 * Run shellscript: $ sh run.sh
	 */

	// invoices.ExamplePdf()
	// invoices.InitInvoiceOutput()
	server.Start()
	// if err := server.Uploader("pdf.pdf"); err != nil {
	// 	log.Fatalln(err)
	// }
}
