package invoices

import "github.com/jung-kurt/gofpdf"

// PDFData -
type PDFData struct {
	Data      []*PDFData
	Headlines []string
}

// Lines -
type Lines struct {
	InvoiceNum int
	Name       string
}

// WritePDF -
func WritePDF(pdfData *PDFData) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	_ = pdf
}

// WritePDFHeader -
func WritePDFHeader(h []string) {
	for _, val := range h {
		_ = val
	}
}

func (d *PDFData) setPDFData() {

}
