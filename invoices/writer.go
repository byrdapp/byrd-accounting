package invoices

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/byblix/byrd-accounting/storage"
	"github.com/jung-kurt/gofpdf"
)

const (
	denmark                = "Denmark"
	danmark                = "Danmark"
	other                  = "Other"
	dkk                    = "DKK"
	eur                    = "EUR"
	productLineNumber      = 2
	photographerCut        = 15
	unlimitedAmountCredits = 0
	euroToDKKPrice         = 7.25
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

var ftrHdrSizes = []float64{20, 30, 50, 20, 30, 30, 20, 40}

// WriteInvoicesPDF (abstraction) creates PDF from data
func WriteInvoicesPDF(invoices []*BookedInvoice) ([]byte, error) {
	db, err := storage.InitFirebaseDB()
	pdfLines := handleValues(db, invoices)
	totals := calcTotalVals(pdfLines)
	pdf := newPDF()
	pdf = writeHeader(pdf, []string{"Invoice#", "Date", "Customer", "Country", "Max seller cut", "Min. Byrd cut", "VAT", "Total price (DKK)"})
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

func handleValues(db *storage.DBInstance, invoices []*BookedInvoice) []*PDFLine {
	pdfLines := []*PDFLine{}
	// totalVals := []*TotalVals{}
	for _, invoice := range invoices {
		for _, line := range invoice.Lines {
			if line.LineNumber == productLineNumber {
				product, err := storage.GetSubscriptionProducts(db, line.getProductNum())
				if err != nil {
					log.Panicf("Didnt get products from FB: %s", err)
				}
				pdfLine := &PDFLine{
					InvoiceNum:   invoice.BookedInvoiceNumber,
					Recipient:    invoice.Recipient,
					Date:         invoice.Date,
					MaxSellerCut: invoice.maxSellerCut(product),
					MinByrdInc:   invoice.minByrdInc(product),
					VAT:          invoice.applyTax(),
					NetAmount:    invoice.netAmount(),
					Period:       product.Period,
				}
				fmt.Printf("Credits: %v. VAT: %v. Period: %s \n", product.Credits, pdfLine.VAT, pdfLine.Period)
				pdfLines = append(pdfLines, pdfLine)
			}
		}
	}
	return pdfLines
}

// TODO:
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

func writeBody(pdf *gofpdf.Fpdf, pdfLines []*PDFLine) *gofpdf.Fpdf {
	pdf.SetFont("Times", "", 10)
	pdf.SetFillColor(255, 255, 255)
	// {20, 30, 50, 20, 30, 30, 20, 40}
	for _, line := range pdfLines {
		pdf.Cell(20, 10, strconv.Itoa(line.InvoiceNum))
		pdf.Cell(30, 10, line.Date)
		pdf.Cell(50, 10, line.Recipient.Name)
		pdf.Cell(20, 10, line.Recipient.Country)
		pdf.Cell(30, 10, formatFloatToStr(line.MaxSellerCut))
		pdf.Cell(30, 10, formatFloatToStr(line.MinByrdInc))
		pdf.Cell(20, 10, formatFloatToStr(line.VAT))
		pdf.Cell(40, 10, formatFloatToStr(line.NetAmount+line.VAT))
		pdf.Ln(6)
		fmt.Printf("Wrote invoice#: %v to customer: %s\n", line.InvoiceNum, line.Recipient.Name)
	}
	return pdf
}

func writeFooter(pdf *gofpdf.Fpdf, ftr *TotalVals) *gofpdf.Fpdf {
	pdf.SetFont("Times", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.Cell(20, 10, "Total values:")
	pdf.Cell(30, 10, "")
	pdf.Cell(50, 10, "")
	pdf.Cell(20, 10, "")
	pdf.Cell(30, 10, formatFloatToStr(ftr.TotalSellerCut))
	pdf.Cell(30, 10, formatFloatToStr(ftr.TotalByrdInc))
	pdf.Cell(20, 10, formatFloatToStr(ftr.TotalVAT))
	pdf.Cell(40, 10, formatFloatToStr(ftr.TotalNetAmount+ftr.TotalVAT))
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
	if i.Currency != dkk {
		i.NetAmount *= euroToDKKPrice
	}
	return i.NetAmount / parseIntToFloat(s.Credits)
}

func (i *BookedInvoice) minByrdInc(s *storage.SubscriptionProduct) float64 {
	if s.Credits != unlimitedAmountCredits {
		if i.Currency != dkk {
			i.NetAmount *= euroToDKKPrice
		}
		return i.NetAmount - i.maxSellerCut(s)
	}
	return 0
}

func (i *BookedInvoice) maxSellerCut(s *storage.SubscriptionProduct) float64 {
	if s.Credits != unlimitedAmountCredits {
		if i.Currency != dkk {
			i.NetAmount *= euroToDKKPrice
		}
		return photographerCut * i.perCreditPrice(s)
	}
	return 0
}

func (i *BookedInvoice) netAmount() float64 {
	if i.Currency != dkk {
		i.NetAmount *= euroToDKKPrice
	}
	return i.NetAmount
}

func (i *BookedInvoice) applyTax() float64 {
	if i.Recipient.Country == denmark || i.Recipient.Country == danmark {
		return i.VatAmount
	}
	return 0
}

func formatFloatToStr(n float64) string {
	return strconv.FormatFloat(n, 'f', 2, 64)
}

func parseIntToFloat(n int) float64 {
	return float64(n)
}
