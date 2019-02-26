package invoices

// InitInvoiceOutput outputs the PDF as a []byte
func InitInvoiceOutput(d *DateRange) ([]byte, error) {
	/*test*/
	// d := &DateRange{
	// 	From: "2018-12-01",
	// 	To:   "2019-01-1",
	// }
	// d.Query = "date$gte:" + d.From + "$and:date$lte:" + d.To
	/*test*/

	// Set current dates and GET the booked Eco invoices
	invoices, err := getEconomicsBookedInvoices(d.Query)
	if err != nil {
		return nil, err
	}

	// For each invoices (*Collection), fetch the corresponding specific invoice line /invoices/booked/{number}
	specificInvoices, err := getSpecificEcoBookedInvoices(invoices.Collection)
	if err != nil {
		return nil, err
	}

	// Write the invoice
	file, err := WriteInvoicesPDF(specificInvoices)
	if err != nil {
		return nil, err
	}
	return file, nil
}
