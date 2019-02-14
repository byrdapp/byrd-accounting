package pdfwriter

import "github.com/jung-kurt/gofpdf"

type PDF struct {
}

type PDFData struct {
}

type PDFHeadline struct {
}

func WritePDF() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	_ = pdf

}

func WritePDFHeader(h *PDFHeadline) {

}

func (h *PDFHeadline) SetHeadline() {

}
