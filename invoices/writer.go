package invoices

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/byblix/byrd-accounting/storage"
	"github.com/jung-kurt/gofpdf"
	"github.com/leekchan/accounting"
)

const (
	denmark                = "Denmark"
	danmark                = "Danmark"
	other                  = "Other"
	dkk                    = "DKK"
	eur                    = "EUR"
	month                  = "month"
	year                   = "year"
	productLineNumber      = 1
	photographerCut        = 15
	unlimitedAmountCredits = 0
	euroToDKKPrice         = 7.425
)

// PDFLine -
type PDFLine struct {
	InvoiceNum                               int
	Recipient                                *Recipient
	Date, Period                             string
	MaxSellerCut, MinByrdInc, NetAmount, VAT float64
}

// TotalVals -
type TotalVals struct {
	TotalSellerCut, TotalByrdInc, TotalNetAmount, TotalVAT float64
}

var ftrHdrSizes = []float64{20, 20, 50, 20, 20, 30, 30, 30, 30}

// WriteInvoicesPDF (abstraction) creates PDF from data
func WriteInvoicesPDF(invoices []*BookedInvoice) ([]byte, error) {
	ac := &accounting.Accounting{Precision: 2, Thousand: ".", Decimal: ","}
	// Init DB
	db, err := storage.InitFirebaseDB()
	// Gather PDF values in struct
	pdfLines := handleValues(db, invoices)
	// calculate total values
	totals := calcTotalVals(pdfLines)
	// Write new PDF
	pdf := newPDF()
	pdf = writeHeader(pdf, []string{"Inv.#", "Date", "Customer", "Country", "Period", "Max seller cut", "Min. Byrd cut", "VAT", "Total price (DKK)"})
	pdf = writeBody(pdf, pdfLines, ac)
	pdf = writeFooter(pdf, totals, ac)
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
	bImg := storage.GetAWSSecrets("byrd.png")
	r := bytes.NewReader(bImg)
	opts := gofpdf.ImageOptions{
		ImageType: "PNG",
		ReadDpi:   true,
	}
	pdf.RegisterImageOptionsReader("byrd.png", opts, r)
	pdf.ImageOptions("byrd.png", 225, 5, 25, 25, false, opts, 0, "")
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

func handleValues(db *storage.DBInstance, invoices []*BookedInvoice) []*PDFLine {
	pdfLines := []*PDFLine{}
	// totalVals := []*TotalVals{}
	for _, invoice := range invoices {
		for _, line := range invoice.Lines {
			if line.LineNumber == productLineNumber {
				product, err := storage.GetSubscriptionProducts(db, line.Product.ProductNumber)
				if err != nil {
					log.Panicf("Didnt get products from FB: %s", err)
				}
				pdfLine := &PDFLine{
					InvoiceNum:   invoice.BookedInvoiceNumber,
					Recipient:    invoice.Recipient,
					Date:         invoice.Date,
					MaxSellerCut: maxSellerCut(product),
					MinByrdInc:   invoice.minByrdInc(line, product),
					Period:       setPeriod(product.Period),
					VAT:          invoice.applyTax(line),
					NetAmount:    invoice.netAmount(line),
				}
				fmt.Printf("Credits: %v. VAT: %v. Period: %s \n", product.Credits, pdfLine.VAT, pdfLine.Period)
				pdfLines = append(pdfLines, pdfLine)
			}
		}
	}
	return pdfLines
}

func calcTotalVals(vals []*PDFLine) *TotalVals {
	totalVals := &TotalVals{}
	for _, v := range vals {
		totalVals.TotalByrdInc += v.MinByrdInc
		totalVals.TotalNetAmount += v.NetAmount
		totalVals.TotalSellerCut += v.MaxSellerCut
		totalVals.TotalVAT += v.VAT
	}
	return totalVals
}

func writeBody(pdf *gofpdf.Fpdf, pdfLines []*PDFLine, ac *accounting.Accounting) *gofpdf.Fpdf {
	pdf.SetFont("Times", "", 10)

	pdf.SetFillColor(240, 240, 240)
	// {20, 30, 50, 20, 30, 30, 20, 40}
	for _, line := range pdfLines {
		pdf.CellFormat(20, 8, strconv.Itoa(line.InvoiceNum), "1", 0, "", true, 0, "")
		pdf.CellFormat(20, 8, line.Date, "1", 0, "", true, 0, "")
		pdf.CellFormat(50, 8, line.Recipient.Name, "1", 0, "", true, 0, "")
		pdf.CellFormat(20, 8, line.Recipient.Country, "1", 0, "", true, 0, "")
		pdf.CellFormat(20, 8, line.Period, "1", 0, "", true, 0, "")
		pdf.CellFormat(30, 8, ac.FormatMoneyFloat64(line.MaxSellerCut), "1", 0, "", true, 0, "")
		pdf.CellFormat(30, 8, ac.FormatMoneyFloat64(line.MinByrdInc), "1", 0, "", true, 0, "")
		pdf.CellFormat(30, 8, ac.FormatMoneyFloat64(line.VAT), "1", 0, "", true, 0, "")
		pdf.CellFormat(30, 8, ac.FormatMoneyFloat64(line.NetAmount+line.VAT), "1", 0, "", true, 0, "")
		pdf.Ln(6)
		fmt.Printf("Wrote invoice#: %v to customer: %s with amount: %v\n", line.InvoiceNum, line.Recipient.Name, line.NetAmount)
	}
	return pdf
}

func writeFooter(pdf *gofpdf.Fpdf, ftr *TotalVals, ac *accounting.Accounting) *gofpdf.Fpdf {
	pdf.SetFont("Times", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.Cell(20, 10, "Total amounts:")
	pdf.Cell(20, 10, "")
	pdf.Cell(50, 10, "")
	pdf.Cell(20, 10, "")
	pdf.Cell(20, 10, "")
	pdf.Cell(30, 10, ac.FormatMoneyFloat64(ftr.TotalSellerCut))
	pdf.Cell(30, 10, ac.FormatMoneyFloat64(ftr.TotalByrdInc))
	pdf.Cell(30, 10, ac.FormatMoneyFloat64(ftr.TotalVAT))
	pdf.Cell(30, 10, ac.FormatMoneyFloat64(ftr.TotalNetAmount+ftr.TotalVAT))
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

func (i *BookedInvoice) minByrdInc(l *Lines, p *storage.SubscriptionProduct) float64 {
	if p.Credits != unlimitedAmountCredits {
		return l.TotalNetAmount - maxSellerCut(p)
	}
	return 0
}

func maxSellerCut(p *storage.SubscriptionProduct) float64 {
	if p.Credits != unlimitedAmountCredits {
		if p.Period == year {
			p.Credits = calcYearCreditAmount(p)
		}
		creditDKKPrice := photographerCut * euroToDKKPrice
		creditsAmt := parseIntToFloat(p.Credits)
		return creditDKKPrice * creditsAmt
	}
	return 0
}

func calcYearCreditAmount(p *storage.SubscriptionProduct) int {
	return p.Credits * 12
}

func (i *BookedInvoice) netAmount(l *Lines) float64 {
	if i.Currency != dkk {
		l.TotalNetAmount *= euroToDKKPrice
	}
	return l.TotalNetAmount
}

func setPeriod(p string) string {
	switch p {
	case month:
		return "MONTH"
	case year:
		return "YEAR"
	default:
		return "%"
	}
}

func (i *BookedInvoice) applyTax(l *Lines) float64 {
	if i.Recipient.Country == denmark || i.Recipient.Country == danmark {
		return l.VatAmount
	}
	return 0.00
}

func formatFloatToStr(n float64) string {
	return strconv.FormatFloat(n, 'f', 2, 64)
}

func parseIntToFloat(n int) float64 {
	return float64(n)
}
