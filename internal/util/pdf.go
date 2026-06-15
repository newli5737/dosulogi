package util

import (
	"bytes"
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/image/webp"
)

type InvoicePDFData struct {
	Code       string
	Serial     string
	Template   string
	IssuedDate string
	DueDate    string

	CompanyName    string
	CompanyTaxCode string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string
	CompanyTagline string

	CustomerName    string
	CustomerTaxCode string
	CustomerAddress string
	CustomerPhone   string

	Route         string
	VehiclePlate  string
	ShipmentCode  string

	Items     []InvoiceItem
	Subtotal  float64
	TaxRate   float64
	TaxAmount float64
	Total     float64
	Currency  string
}

type InvoiceItem struct {
	Description string
	Unit        string
	Qty         float64
	UnitPrice   float64
	Amount      float64
	Route       string
	Plate       string
}

var fontDir string
var logoPNGPath string

func initAssets() error {
	if fontDir != "" {
		return nil
	}
	base, err := os.Getwd()
	if err != nil {
		return err
	}
	candidates := []string{
		filepath.Join(base, "internal", "util"),
		filepath.Join(base, "..", "internal", "util"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "fonts", "NotoSans-Regular.ttf")); err == nil {
			fontDir = filepath.Join(c, "fonts")
			logoPNGPath = filepath.Join(c, "assets", "logo.png")
			break
		}
	}
	if fontDir == "" {
		return fmt.Errorf("không tìm thấy font NotoSans")
	}
	if err := ensureLogoPNG(filepath.Join(filepath.Dir(fontDir), "assets")); err != nil {
		return err
	}
	return nil
}

func ensureLogoPNG(assetsDir string) error {
	pngPath := filepath.Join(assetsDir, "logo.png")
	if _, err := os.Stat(pngPath); err == nil {
		logoPNGPath = pngPath
		return nil
	}
	webpPath := filepath.Join(assetsDir, "logo.webp")
	raw, err := os.ReadFile(webpPath)
	if err != nil {
		return nil
	}
	img, err := webp.Decode(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	f, err := os.Create(pngPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return err
	}
	logoPNGPath = pngPath
	return nil
}

func GenerateInvoicePDF(data InvoicePDFData, outputDir, filename string) (string, error) {
	if err := initAssets(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.AddPage()

	pdf.AddUTF8Font("Noto", "", filepath.Join(fontDir, "NotoSans-Regular.ttf"))
	pdf.AddUTF8Font("Noto", "B", filepath.Join(fontDir, "NotoSans-Bold.ttf"))

	// Logo + company block
	if logoPNGPath != "" {
		if _, err := os.Stat(logoPNGPath); err == nil {
			pdf.ImageOptions(logoPNGPath, 12, 10, 28, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		}
	}
	pdf.SetFont("Noto", "B", 11)
	pdf.SetXY(42, 10)
	pdf.Cell(0, 5, data.CompanyName)
	pdf.SetFont("Noto", "", 8)
	pdf.SetXY(42, 16)
	if data.CompanyTagline != "" {
		pdf.MultiCell(156, 3.5, data.CompanyTagline, "", "L", false)
	}
	pdf.SetFont("Noto", "", 9)
	pdf.SetXY(42, pdf.GetY()+1)
	pdf.MultiCell(156, 4, fmt.Sprintf("MST: %s\nĐịa chỉ: %s\nĐT: %s  Email: %s",
		orDash(data.CompanyTaxCode), data.CompanyAddress, data.CompanyPhone, data.CompanyEmail), "", "L", false)

	// National header
	pdf.SetY(pdf.GetY() + 4)
	pdf.SetFont("Noto", "B", 10)
	pdf.CellFormat(0, 5, "CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM", "", 1, "C", false, 0, "")
	pdf.SetFont("Noto", "", 9)
	pdf.CellFormat(0, 5, "Độc lập - Tự do - Hạnh phúc", "", 1, "C", false, 0, "")
	pdf.Ln(2)
	pdf.SetFont("Noto", "B", 14)
	pdf.CellFormat(0, 8, "HÓA ĐƠN GIÁ TRỊ GIA TĂNG", "", 1, "C", false, 0, "")
	pdf.SetFont("Noto", "", 10)
	pdf.CellFormat(0, 6, "(Dịch vụ vận chuyển hàng hóa)", "", 1, "C", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Noto", "", 9)
	serial := data.Serial
	if serial == "" {
		serial = "DL/26E"
	}
	template := data.Template
	if template == "" {
		template = "1/001"
	}
	pdf.CellFormat(95, 5, fmt.Sprintf("Mẫu số: %s", template), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 5, fmt.Sprintf("Ký hiệu: %s", serial), "", 1, "L", false, 0, "")
	pdf.CellFormat(95, 5, fmt.Sprintf("Số: %s", data.Code), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 5, fmt.Sprintf("Ngày %s", formatVietnameseDate(data.IssuedDate)), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	drawPartyBox(pdf, "Đơn vị bán hàng", data.CompanyName, data.CompanyTaxCode, data.CompanyAddress, data.CompanyPhone)
	drawPartyBox(pdf, "Đơn vị mua hàng", data.CustomerName, data.CustomerTaxCode, data.CustomerAddress, data.CustomerPhone)

	if data.ShipmentCode != "" || data.Route != "" {
		pdf.SetFont("Noto", "", 9)
		pdf.CellFormat(0, 5, fmt.Sprintf("Mã vận đơn: %s   Hành trình: %s   Biển kiểm soát: %s",
			data.ShipmentCode, data.Route, data.VehiclePlate), "", 1, "L", false, 0, "")
		pdf.Ln(1)
	}

	// Items table
	headers := []string{"STT", "Tên hàng hóa, dịch vụ", "ĐVT", "SL", "Đơn giá", "Thành tiền"}
	widths := []float64{10, 72, 14, 14, 32, 32}
	pdf.SetFont("Noto", "B", 8)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont("Noto", "", 8)
	for i, item := range data.Items {
		amt := item.Amount
		if amt == 0 {
			amt = item.Qty * item.UnitPrice
		}
		unit := item.Unit
		if unit == "" {
			unit = "Chuyến"
		}
		desc := item.Description
		if item.Route != "" {
			desc += " | HT: " + item.Route
		}
		if item.Plate != "" {
			desc += " | BSX: " + item.Plate
		}
		pdf.CellFormat(widths[0], 7, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[1], 7, desc, "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[2], 7, unit, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[3], 7, formatQty(item.Qty), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[4], 7, formatMoney(item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[5], 7, formatMoney(amt), "1", 1, "R", false, 0, "")
	}

	pdf.Ln(2)
	pdf.SetFont("Noto", "", 9)
	pdf.CellFormat(130, 6, "Cộng tiền hàng (chưa thuế GTGT):", "", 0, "R", false, 0, "")
	pdf.CellFormat(44, 6, formatMoney(data.Subtotal)+" "+data.Currency, "", 1, "R", false, 0, "")
	pdf.CellFormat(130, 6, fmt.Sprintf("Thuế suất GTGT: %.0f%%", data.TaxRate), "", 0, "R", false, 0, "")
	pdf.CellFormat(44, 6, formatMoney(data.TaxAmount)+" "+data.Currency, "", 1, "R", false, 0, "")
	pdf.SetFont("Noto", "B", 10)
	pdf.CellFormat(130, 7, "Tổng cộng tiền thanh toán:", "", 0, "R", false, 0, "")
	pdf.CellFormat(44, 7, formatMoney(data.Total)+" "+data.Currency, "", 1, "R", false, 0, "")
	pdf.SetFont("Noto", "", 9)
	pdf.MultiCell(0, 5, "Số tiền bằng chữ: "+numberToVietnameseWords(data.Total), "", "L", false)
	if data.DueDate != "" {
		pdf.CellFormat(0, 5, fmt.Sprintf("Hạn thanh toán: %s", formatVietnameseDate(data.DueDate)), "", 1, "L", false, 0, "")
	}
	pdf.Ln(6)

	pdf.SetFont("Noto", "", 9)
	pdf.CellFormat(95, 5, "Người mua hàng", "", 0, "C", false, 0, "")
	pdf.CellFormat(95, 5, "Người bán hàng", "", 1, "C", false, 0, "")
	pdf.SetFont("Noto", "", 8)
	pdf.CellFormat(95, 5, "(Ký, ghi rõ họ tên)", "", 0, "C", false, 0, "")
	pdf.CellFormat(95, 5, "(Ký, đóng dấu, ghi rõ họ tên)", "", 1, "C", false, 0, "")

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

func drawPartyBox(pdf *gofpdf.Fpdf, title, name, tax, address, phone string) {
	pdf.SetFont("Noto", "B", 9)
	pdf.CellFormat(0, 5, title, "", 1, "L", false, 0, "")
	pdf.SetFont("Noto", "", 9)
	if name == "" {
		name = "—"
	}
	pdf.MultiCell(0, 4, fmt.Sprintf("Tên: %s\nMST: %s\nĐịa chỉ: %s\nĐT: %s",
		name, orDash(tax), orDash(address), orDash(phone)), "", "L", false)
	pdf.Ln(1)
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func formatMoney(v float64) string {
	n := int64(math.Round(v))
	s := fmt.Sprintf("%d", n)
	if n < 0 {
		s = s[1:]
	}
	var out []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, byte(c))
	}
	if n < 0 {
		return "-" + string(out)
	}
	return string(out)
}

func formatQty(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.2f", v)
}

func formatVietnameseDate(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return fmt.Sprintf("ngày %02d tháng %02d năm %d", t.Day(), int(t.Month()), t.Year())
}

func readTriple(num int, digits, teen []string) string {
	h, t, u := num/100, (num/10)%10, num%10
	var p []string
	if h > 0 {
		p = append(p, digits[h]+" trăm")
	}
	if t > 1 {
		p = append(p, digits[t]+" mươi")
		if u == 1 {
			p = append(p, "mốt")
		} else if u == 5 {
			p = append(p, "lăm")
		} else if u > 0 {
			p = append(p, digits[u])
		}
	} else if t == 1 {
		p = append(p, teen[u])
	} else if u > 0 {
		if h > 0 {
			p = append(p, "lẻ")
		}
		p = append(p, digits[u])
	}
	return strings.Join(p, " ")
}

func numberToVietnameseWords(n float64) string {
	amount := int64(math.Round(n))
	if amount == 0 {
		return "Không đồng"
	}
	units := []string{"", " nghìn", " triệu", " tỷ"}
	digits := []string{"không", "một", "hai", "ba", "bốn", "năm", "sáu", "bảy", "tám", "chín"}
	teen := []string{"mười", "mười một", "mười hai", "mười ba", "mười bốn", "mười lăm", "mười sáu", "mười bảy", "mười tám", "mười chín"}

	var parts []string
	idx := 0
	for amount > 0 && idx < len(units) {
		chunk := int(amount % 1000)
		if chunk > 0 {
			text := readTriple(chunk, digits, teen) + units[idx]
			parts = append([]string{text}, parts...)
		}
		amount /= 1000
		idx++
	}
	out := strings.Join(parts, " ")
	if out == "" {
		out = "Không"
	}
	return strings.ToUpper(out[:1]) + out[1:] + " đồng"
}

func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
