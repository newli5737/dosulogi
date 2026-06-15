package util

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type InvoicePDFData struct {
	Code       string
	Customer   string
	Items      []InvoiceItem
	Subtotal   float64
	TaxRate    float64
	TaxAmount  float64
	Total      float64
	Currency   string
	DueDate    string
	IssuedDate string
}

type InvoiceItem struct {
	Description string
	Qty         float64
	UnitPrice   float64
	Amount      float64
}

func GenerateInvoicePDF(data InvoicePDFData, outputDir, filename string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "INVOICE")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(40, 8, fmt.Sprintf("Code: %s", data.Code))
	pdf.Ln(6)
	pdf.Cell(40, 8, fmt.Sprintf("Customer: %s", data.Customer))
	pdf.Ln(6)
	pdf.Cell(40, 8, fmt.Sprintf("Issued: %s  Due: %s", data.IssuedDate, data.DueDate))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(80, 8, "Description", "1", 0, "L", false, 0, "")
	pdf.CellFormat(25, 8, "Qty", "1", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, "Unit Price", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, "Amount", "1", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	for _, item := range data.Items {
		pdf.CellFormat(80, 8, item.Description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%.2f", item.Qty), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 8, fmt.Sprintf("%.2f", item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", item.Amount), "1", 1, "R", false, 0, "")
	}

	pdf.Ln(4)
	pdf.CellFormat(140, 8, "Subtotal", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f %s", data.Subtotal, data.Currency), "", 1, "R", false, 0, "")
	pdf.CellFormat(140, 8, fmt.Sprintf("Tax (%.1f%%)", data.TaxRate), "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f %s", data.TaxAmount, data.Currency), "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(140, 8, "Total", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f %s", data.Total, data.Currency), "", 1, "R", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", err
	}

	fullPath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(fullPath, buf.Bytes(), 0644); err != nil {
		return "", err
	}
	return fullPath, nil
}

func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
