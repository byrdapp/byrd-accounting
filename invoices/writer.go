package invoices

import (
	"fmt"
	"strconv"
)

const (
	denmark = "Denmark"
	other   = "Other"
)

// PDFLines -
type PDFLines struct {
	InvoiceNum              int
	Recipient               *Recipient
	Date                    string
	PotentialCreditOutbound float64
	PotentialAmountOutbound float64
	ByrdInc                 float64
	VAT                     float64
}

// WritePDF is an abstraction of the loop with real data
func WritePDF(lines []*Lines, invoice *BookedInvoice) error {
	for idx, line := range lines {
		pdfLines := &PDFLines{
			InvoiceNum:              invoice.BookedInvoiceNumber,
			Recipient:               invoice.Recipient,
			Date:                    invoice.Date,
			PotentialCreditOutbound: line.potentialCreditOutbound(invoice),
			PotentialAmountOutbound: line.potentialEuroAmountOutbound(invoice),
			ByrdInc:                 line.byrdIncome(invoice),
			VAT:                     line.applyTax(invoice),
		}
		fmt.Printf("Wrote line %s\n", strconv.Itoa(idx))
		fmt.Printf("%+v\n", pdfLines)
	}
	return nil
}

// WritePDFHeader -
func writePDFHeader() {
	headlines := make(map[string]string)
	for _, val := range headlines {
		_ = val
	}
}

func (v *Lines) potentialCreditOutbound(i *BookedInvoice) float64 {
	return i.NetAmount / v.CreditQuantity
}

func (v *Lines) byrdIncome(i *BookedInvoice) float64 {
	return i.NetAmount - v.potentialEuroAmountOutbound(i)
}

func (v *Lines) potentialEuroAmountOutbound(i *BookedInvoice) float64 {
	return photographerCut * v.CreditQuantity
}

func (v *Lines) applyTax(i *BookedInvoice) float64 {
	if i.Recipient.Country == denmark {
		return i.VatAmount
	}
	return 0
}
