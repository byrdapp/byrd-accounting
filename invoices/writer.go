package invoices

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
)

const (
	denmark           = "Denmark"
	other             = "Other"
	productLineNumber = 2
	photographerCut   = 15
)

// PDFLines -
type PDFLines struct {
	InvoiceNum     int
	Recipient      *Recipient
	Date           string
	MaxSellerCut   float64
	MinByrdInc     float64
	TotalNetAmount float64
	VAT            float64
}

// WriteInvoicesPDF (abstraction) creates PDF from data
func WriteInvoicesPDF(invoices []*BookedInvoice) ([]byte, error) {
	pdfLines := destructValues(invoices)
	pdf := newPDF()
	pdf = writeHeader(pdf, []string{"Invoice#", "Date", "Customer", "Country", "Max seller cut", "Min. Byrd cut", "VAT", "Total price"})
	pdf = writeBody(pdf, pdfLines)
	pdf = writeFooter(pdf)
	// Write footer with page #
	fileName, err := createPDF(pdf)
	if err != nil {
		return nil, err
	}
	fmt.Println("Created PDF")
	return fileName, nil
}

func newPDF() *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "mm", "Letter", "")
	pdf.AddPage()
	pdf.SetFont("Times", "B", 16)
	pdf.Cell(30, 10, "Media usage report")
	pdf.Ln(10)
	pdf.SetFont("Times", "", 10)
	pdf.Cell(30, 10, "Generated: "+time.Now().Format("Mon Jan 2, 2006"))
	// pdf.ImageOptions("byrd.png", 225, 5, 25, 25, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	pdf.Ln(14)
	return pdf
}

func writeHeader(pdf *gofpdf.Fpdf, hdr []string) *gofpdf.Fpdf {
	pdf.SetFont("Times", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	for _, str := range hdr {
		pdf.CellFormat(30, 7, str, "1", 0, "", true, 0, "")
	}
	pdf.Ln(-1)
	return pdf
}

func destructValues(invoices []*BookedInvoice) []*PDFLines {
	pdfLines := []*PDFLines{}
	for _, invoice := range invoices {
		for _, line := range invoice.Lines {
			if line.LineNumber == productLineNumber {
				pdfLine := &PDFLines{
					InvoiceNum:     invoice.BookedInvoiceNumber,
					Recipient:      invoice.Recipient,
					Date:           invoice.Date,
					MaxSellerCut:   line.maxSellerCut(invoice),
					MinByrdInc:     line.minByrdInc(invoice),
					VAT:            line.applyTax(invoice),
					TotalNetAmount: invoice.NetAmount,
				}
				pdfLines = append(pdfLines, pdfLine)
			}
		}
	}
	return pdfLines
}

func writeBody(pdf *gofpdf.Fpdf, pdfLines []*PDFLines) *gofpdf.Fpdf {
	pdf.SetFont("Times", "", 10)
	pdf.SetFillColor(255, 255, 255)
	for _, line := range pdfLines {
		pdf.Cell(30, 10, strconv.Itoa(line.InvoiceNum))
		pdf.Cell(30, 10, line.Date)
		pdf.Cell(30, 10, line.Recipient.Name)
		pdf.Cell(30, 10, line.Recipient.Country)
		pdf.Cell(30, 10, formatFloat(line.MaxSellerCut))
		pdf.Cell(30, 10, formatFloat(line.MinByrdInc))
		pdf.Cell(30, 10, formatFloat(line.VAT))
		pdf.Cell(30, 10, formatFloat(line.TotalNetAmount+line.VAT))
		pdf.Ln(6)
		fmt.Printf("Wrote invoice#: %v to customer: %s\n", line.InvoiceNum, line.Recipient.Name)
	}
	return pdf
}

func writeFooter(pdf *gofpdf.Fpdf) *gofpdf.Fpdf {
	return pdf
}

func createPDF(pdf *gofpdf.Fpdf) ([]byte, error) {
	fmt.Println("Creating mem pdf")

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		if err := pdf.OutputAndClose(pw); err != nil {
			panic(err)
		}
	}()

	b, err := ioutil.ReadAll(pr)
	if err != nil {
		panic(err)
	}
	return b, nil

}

func (v *Lines) perCreditPrice(i *BookedInvoice) float64 {
	return i.NetAmount / v.CreditQuantity
}

func (v *Lines) minByrdInc(i *BookedInvoice) float64 {
	return i.NetAmount - v.maxSellerCut(i)
}

func (v *Lines) maxSellerCut(i *BookedInvoice) float64 {
	return photographerCut * v.perCreditPrice(i)
}

func (v *Lines) applyTax(i *BookedInvoice) float64 {
	if i.Recipient.Country == denmark {
		return i.VatAmount
	}
	return 0
}

func formatFloat(n float64) string {
	return strconv.FormatFloat(n, 'f', 2, 64)
}
