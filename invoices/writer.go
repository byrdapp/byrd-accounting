package invoices

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/byblix/byrd-accounting/storage"
	"github.com/jung-kurt/gofpdf"
)

const (
	denmark                = "Denmark"
	other                  = "Other"
	productLineNumber      = 2
	photographerCut        = 15
	unlimitedAmountCredits = 0
)

// PDFLines -
type PDFLines struct {
	InvoiceNum                               int
	Recipient                                *Recipient
	Date, Period                             string
	MaxSellerCut, MinByrdInc, NetAmount, VAT float64
}

// TotalVals -
type TotalVals struct {
	TotalSellerCut, TotalByrdInc, TotalNetAmount, TotalVAT float64
}

var ftrHdrSizes = []float64{20, 30, 50, 20, 30, 30, 20, 40}

// WriteInvoicesPDF (abstraction) creates PDF from data
func WriteInvoicesPDF(invoices []*BookedInvoice) ([]byte, error) {
	db, err := storage.InitFirebaseDB()
	pdfLines := handleValues(db, invoices)
	totals := calcTotalVals(pdfLines)
	pdf := newPDF()
	pdf = writeHeader(pdf, []string{"Invoice#", "Date", "Customer", "Country", "Max seller cut", "Min. Byrd cut", "VAT", "Total price"})
	pdf = writeBody(pdf, pdfLines)
	pdf = writeFooter(pdf, totals)
	// Write footer with page #
	file, err := createPDF(pdf)
	if err != nil {
		return nil, err
	}
	fmt.Println("Created PDF")
	return file, nil
}

func newPDF() *gofpdf.Fpdf {
	pdf := gofpdf.New("L", "mm", "Letter", "")
	pdf.AddPage()
	pdf.SetFont("Times", "B", 16)
	pdf.Cell(40, 10, "Media usage report")
	pdf.Ln(10)
	pdf.SetFont("Times", "", 10)
	pdf.Cell(40, 10, "Generated: "+time.Now().Format("Mon Jan 2, 2006"))
	// bImg := storage.GetAWSSecrets("byrd.png")
	// img := pdf.RegisterImageOptionsReader()
	pdf.ImageOptions("byrd.png", 225, 5, 25, 25, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	pdf.Ln(14)
	return pdf
}

func writeHeader(pdf *gofpdf.Fpdf, hdr []string) *gofpdf.Fpdf {
	pdf.SetFont("Times", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	for i, str := range hdr {
		pdf.CellFormat(ftrHdrSizes[i], 7, str, "1", 0, "", true, 0, "")
	}
	pdf.Ln(-1)
	return pdf
}

func handleValues(db *storage.DBInstance, invoices []*BookedInvoice) []*PDFLines {
	pdfLines := []*PDFLines{}
	// totalVals := []*TotalVals{}
	for _, invoice := range invoices {
		for _, line := range invoice.Lines {
			if line.LineNumber == productLineNumber {
				product, err := storage.GetSubscriptionProducts(db, line.getProductNum())
				if err != nil {
					log.Panicf("Didnt get products from FB: %s", err)
				}
				pdfLine := &PDFLines{
					InvoiceNum:   invoice.BookedInvoiceNumber,
					Recipient:    invoice.Recipient,
					Date:         invoice.Date,
					MaxSellerCut: invoice.maxSellerCut(product),
					MinByrdInc:   line.minByrdInc(invoice, product),
					VAT:          applyTax(invoice),
					NetAmount:    invoice.NetAmount,
					Period:       product.Period,
				}
				if product.Credits == unlimitedAmountCredits {
					val, _ := strconv.ParseFloat("Unltd.", 64)
					pdfLine.MaxSellerCut = val
					pdfLine.MinByrdInc = val
				}
				spew.Printf("Credits: %v\n", pdfLine.VAT)
				pdfLines = append(pdfLines, pdfLine)
			}
		}
	}
	return pdfLines
}

// TODO:
func calcTotalVals([]*PDFLines) *TotalVals {
	return nil
}

// TODO:
func parseEuroToDKK(euro float64) float64 {
	return 0
}

func writeBody(pdf *gofpdf.Fpdf, pdfLines []*PDFLines) *gofpdf.Fpdf {
	pdf.SetFont("Times", "", 10)
	pdf.SetFillColor(255, 255, 255)
	// {20, 30, 50, 20, 30, 30, 20, 40}
	for _, line := range pdfLines {
		pdf.Cell(20, 10, strconv.Itoa(line.InvoiceNum))
		pdf.Cell(30, 10, line.Date)
		pdf.Cell(50, 10, line.Recipient.Name)
		pdf.Cell(20, 10, line.Recipient.Country)
		pdf.Cell(30, 10, formatFloat(line.MaxSellerCut))
		pdf.Cell(30, 10, formatFloat(line.MinByrdInc))
		pdf.Cell(20, 10, formatFloat(line.VAT))
		pdf.Cell(40, 10, formatFloat(line.NetAmount+line.VAT))
		pdf.Ln(6)
		fmt.Printf("Wrote invoice#: %v to customer: %s\n", line.InvoiceNum, line.Recipient.Name)
	}
	return pdf
}

func writeFooter(pdf *gofpdf.Fpdf, ftr *TotalVals) *gofpdf.Fpdf {
	pdf.SetFont("Times", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	// pdf.CellFormat(sizes[i], 7, ftr, "1", 0, "", true, 0, "")
	pdf.Ln(-1)
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
		return nil, err
	}
	return b, nil

}

func (v *Lines) getProductNum() string {
	return v.Product.ProductNumber
}

func (i *BookedInvoice) perCreditPrice(s *storage.SubscriptionProduct) float64 {
	return i.NetAmount / parseFloat(s.Credits)
}

func (v *Lines) minByrdInc(i *BookedInvoice, s *storage.SubscriptionProduct) float64 {
	return i.NetAmount - i.maxSellerCut(s)
}

func (i *BookedInvoice) maxSellerCut(s *storage.SubscriptionProduct) float64 {
	return photographerCut * i.perCreditPrice(s)
}

func applyTax(i *BookedInvoice) float64 {
	if i.Recipient.Country == denmark {
		return i.VatAmount
	}
	return 0
}

func formatFloat(n float64) string {
	return strconv.FormatFloat(n, 'f', 2, 64)
}

func parseFloat(n int) float64 {
	return float64(n)
}
