package invoices

// InitInvoiceOutput outputs the PDF as a []byte
func InitInvoiceOutput(d *DateRange) ([]byte, error) {
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
