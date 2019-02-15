package invoices

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jung-kurt/gofpdf"
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

// WriteInvoicesPDF is an abstraction of the loop with real data
func WriteInvoicesPDF(invoice *BookedInvoice, lines []*Lines) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 10)
	pdf.SetTopMargin(30)
	for _, line := range lines {
		if line.LineNumber == creditLineNumber {
			pdfLines := PDFLines{
				InvoiceNum:              invoice.BookedInvoiceNumber,
				Recipient:               invoice.Recipient,
				Date:                    invoice.Date,
				PotentialCreditOutbound: line.potentialCreditOutbound(invoice),
				PotentialAmountOutbound: line.potentialEuroAmountOutbound(invoice),
				ByrdInc:                 line.byrdIncome(invoice),
				VAT:                     line.applyTax(invoice),
			}
			pdfJSON, _ := json.Marshal(pdfLines)
			pdf.Cell(40, 10, string(pdfJSON)+"\n")
			fmt.Printf("%+v\n", pdfLines)
		}
	}
	if err := pdf.OutputFileAndClose("pdf.pdf"); err != nil {
		logger.Fatalln(err)
	}
	return nil
}

// ExamplePdf -
func ExamplePdf() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTopMargin(30)
	pdf.SetHeaderFunc(func() {
		pdf.SetY(5)
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(30, 10, "Byrd Accounting")
		pdf.Ln(20)
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-10)
		pdf.SetFont("Arial", "I", 6)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
	pdf.AliasNbPages("")
	pdf.AddPage()
	pdf.SetFont("Times", "", 12)
	for j := 1; j <= 10; j++ {
		pdf.CellFormat(0, 10, fmt.Sprintf("Lolololo lol olo: %v", j),
			"", 1, "", false, 0, "")
	}
	if err := pdf.OutputFileAndClose("ExamplePDF.pdf"); err != nil {
		log.Fatalln(err)
	}
}

// WritePDFHeader -
func writePDFHeader() {
	headlines := make(map[string]string)
	for _, val := range headlines {
		_ = val
	}
}

func mkdirPDF() {
	if err := os.Mkdir("pdfs", 32); err != nil {
		log.Fatalln(err)
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
