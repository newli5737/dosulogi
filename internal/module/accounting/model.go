package accounting

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID         uuid.UUID       `json:"id"`
	Code       string          `json:"code"`
	CustomerID uuid.UUID       `json:"customer_id"`
	ContractID *uuid.UUID      `json:"contract_id"`
	ShipmentID *uuid.UUID      `json:"shipment_id"`
	Items      json.RawMessage `json:"items"`
	Subtotal   *float64        `json:"subtotal"`
	TaxRate    float64         `json:"tax_rate"`
	TaxAmount  *float64        `json:"tax_amount"`
	Total      *float64        `json:"total"`
	Currency   string          `json:"currency"`
	Status     string          `json:"status"`
	DueDate    *time.Time      `json:"due_date"`
	PaidAt     *time.Time      `json:"paid_at"`
	FileURL    *string         `json:"file_url"`
	CreatedBy  *uuid.UUID      `json:"created_by"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Payment struct {
	ID            uuid.UUID `json:"id"`
	InvoiceID     uuid.UUID `json:"invoice_id"`
	Amount        *float64  `json:"amount"`
	Method        *string   `json:"method"`
	ReferenceCode *string   `json:"reference_code"`
	SePayTxnID    *string   `json:"sepay_txn_id"`
	MatchedAuto   bool      `json:"matched_auto"`
	Note          *string   `json:"note"`
	CreatedAt     time.Time `json:"created_at"`
}

type LineItem struct {
	Description string  `json:"description"`
	Qty         float64 `json:"qty"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
}

type RevenueReport struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
}

type ARReport struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	TotalDue     float64   `json:"total_due"`
	InvoiceCount int       `json:"invoice_count"`
}
