package accounting

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/integration/mailer"
	"github.com/dosu-logi/logistics-erp/internal/integration/sepay"
	"github.com/dosu-logi/logistics-erp/internal/util"
	"github.com/google/uuid"
)

type Service struct {
	repo      *Repository
	mailer    *mailer.Client
	uploadDir string
}

func NewService(repo *Repository, mailer *mailer.Client, uploadDir string) *Service {
	return &Service{repo: repo, mailer: mailer, uploadDir: uploadDir}
}

func calcInvoiceTotals(items []LineItem, taxRate float64) (subtotal, taxAmount, total float64) {
	for _, it := range items {
		amt := it.Qty * it.UnitPrice
		if it.Amount > 0 {
			amt = it.Amount
		}
		subtotal += amt
	}
	taxAmount = subtotal * taxRate / 100
	total = subtotal + taxAmount
	return
}

func (s *Service) CreateInvoice(ctx context.Context, inv *Invoice, items []LineItem) error {
	code, err := s.repo.NextInvoiceCode(ctx)
	if err != nil {
		return err
	}
	inv.Code = code
	if inv.TaxRate == 0 {
		inv.TaxRate = 10
	}
	sub, tax, total := calcInvoiceTotals(items, inv.TaxRate)
	inv.Subtotal = &sub
	inv.TaxAmount = &tax
	inv.Total = &total
	if inv.Status == "" {
		inv.Status = "draft"
	}
	if inv.Currency == "" {
		inv.Currency = "VND"
	}
	itemsJSON, _ := json.Marshal(items)
	inv.Items = itemsJSON
	return s.repo.CreateInvoice(ctx, inv)
}

func (s *Service) ListInvoices(ctx context.Context, status, customerID, from, to string, limit, offset int) ([]Invoice, int, error) {
	return s.repo.ListInvoices(ctx, status, customerID, from, to, limit, offset)
}

func (s *Service) GetInvoice(ctx context.Context, id uuid.UUID) (*Invoice, error) {
	return s.repo.GetInvoice(ctx, id)
}

func (s *Service) UpdateInvoice(ctx context.Context, inv *Invoice, items []LineItem) error {
	if len(items) > 0 {
		sub, tax, total := calcInvoiceTotals(items, inv.TaxRate)
		inv.Subtotal = &sub
		inv.TaxAmount = &tax
		inv.Total = &total
		itemsJSON, _ := json.Marshal(items)
		inv.Items = itemsJSON
	}
	return s.repo.UpdateInvoice(ctx, inv)
}

func (s *Service) SendInvoice(ctx context.Context, id uuid.UUID) error {
	inv, err := s.repo.GetInvoice(ctx, id)
	if err != nil {
		return err
	}
	if err := s.generatePDF(ctx, inv); err != nil {
		return err
	}
	email, _ := s.repo.GetCustomerEmail(ctx, inv.CustomerID)
	if email != "" {
		subject := "Hóa đơn " + inv.Code
		body := "<p>Kính gửi quý khách,</p><p>Đính kèm hóa đơn " + inv.Code + ".</p>"
		_ = s.mailer.SendEmailWithAttachment(ctx, email, subject, body, inv.FileURL)
	}
	inv.Status = "sent"
	return s.repo.UpdateInvoice(ctx, inv)
}

func (s *Service) CancelInvoice(ctx context.Context, id uuid.UUID) error {
	inv, err := s.repo.GetInvoice(ctx, id)
	if err != nil {
		return err
	}
	inv.Status = "cancelled"
	return s.repo.UpdateInvoice(ctx, inv)
}

func (s *Service) generatePDF(ctx context.Context, inv *Invoice) error {
	items, _ := ParseItems(inv.Items)
	name, _ := s.repo.GetCustomerName(ctx, inv.CustomerID)
	pdfItems := make([]util.InvoiceItem, len(items))
	for i, it := range items {
		pdfItems[i] = util.InvoiceItem{Description: it.Description, Qty: it.Qty, UnitPrice: it.UnitPrice, Amount: it.Amount}
	}
	filename := inv.Code + ".pdf"
	path, err := util.GenerateInvoicePDF(util.InvoicePDFData{
		Code: inv.Code, Customer: name, Items: pdfItems,
		Subtotal: *inv.Subtotal, TaxRate: inv.TaxRate, TaxAmount: *inv.TaxAmount,
		Total: *inv.Total, Currency: inv.Currency,
		DueDate: util.FormatDate(inv.DueDate), IssuedDate: inv.CreatedAt.Format("2006-01-02"),
	}, filepath.Join(s.uploadDir, "invoices"), filename)
	if err != nil {
		return err
	}
	inv.FileURL = &path
	return s.repo.UpdateInvoice(ctx, inv)
}

func (s *Service) DownloadInvoice(ctx context.Context, id uuid.UUID) (string, error) {
	inv, err := s.repo.GetInvoice(ctx, id)
	if err != nil {
		return "", err
	}
	if inv.FileURL == nil || *inv.FileURL == "" {
		if err := s.generatePDF(ctx, inv); err != nil {
			return "", err
		}
		inv, _ = s.repo.GetInvoice(ctx, id)
	}
	return *inv.FileURL, nil
}

func (s *Service) CreatePayment(ctx context.Context, p *Payment) error {
	method := "bank_transfer"
	if p.Method == nil {
		p.Method = &method
	}
	if err := s.repo.CreatePayment(ctx, p); err != nil {
		return err
	}
	inv, err := s.repo.GetInvoice(ctx, p.InvoiceID)
	if err != nil {
		return err
	}
	now := time.Now()
	inv.Status = "paid"
	inv.PaidAt = &now
	return s.repo.UpdateInvoice(ctx, inv)
}

func (s *Service) ListPayments(ctx context.Context, limit, offset int) ([]Payment, int, error) {
	return s.repo.ListPayments(ctx, limit, offset)
}

func (s *Service) HandleSePayWebhook(ctx context.Context, payload sepay.WebhookPayload) error {
	inv, err := s.repo.GetInvoiceByCode(ctx, payload.Code)
	if err != nil {
		raw, _ := json.Marshal(payload)
		return s.repo.SaveUnmatchedPayment(ctx, payload.ID, payload.TransferAmount, payload.Code, raw)
	}
	if inv.Total != nil && payload.TransferAmount < *inv.Total {
		raw, _ := json.Marshal(payload)
		return s.repo.SaveUnmatchedPayment(ctx, payload.ID, payload.TransferAmount, payload.Code, raw)
	}
	method := "bank_transfer"
	ref := payload.ReferenceCode
	txnID := payload.ID
	amt := payload.TransferAmount
	p := &Payment{InvoiceID: inv.ID, Amount: &amt, Method: &method, ReferenceCode: &ref, SePayTxnID: &txnID, MatchedAuto: true}
	if err := s.repo.CreatePayment(ctx, p); err != nil {
		return err
	}
	now := time.Now()
	inv.Status = "paid"
	inv.PaidAt = &now
	if err := s.repo.UpdateInvoice(ctx, inv); err != nil {
		return err
	}
	email, _ := s.repo.GetCustomerEmail(ctx, inv.CustomerID)
	if email != "" {
		subject := fmt.Sprintf("Xác nhận thanh toán %s", inv.Code)
		body := fmt.Sprintf("<p>Chúng tôi đã nhận thanh toán %s với số tiền %.0f VND.</p>", inv.Code, payload.TransferAmount)
		_ = s.mailer.SendEmail(ctx, email, subject, body)
	}
	return nil
}

func (s *Service) RevenueReport(ctx context.Context, from, to, groupBy string) ([]RevenueReport, error) {
	return s.repo.RevenueReport(ctx, from, to, groupBy)
}

func (s *Service) ARReport(ctx context.Context) ([]ARReport, error) {
	return s.repo.ARReport(ctx)
}

func (s *Service) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]Invoice, error) {
	return s.repo.ListByCustomer(ctx, customerID)
}
