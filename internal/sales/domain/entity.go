package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LineItem struct {
	Description string  `json:"description"`
	Qty         float64 `json:"qty"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
}

type Opportunity struct {
	ID            uuid.UUID  `json:"id"`
	CustomerID    uuid.UUID  `json:"customer_id"`
	Title         string     `json:"title"`
	Stage         string     `json:"stage"`
	Value         *float64   `json:"value"`
	Currency      string     `json:"currency"`
	ExpectedClose *time.Time `json:"expected_close"`
	AssignedTo    *uuid.UUID `json:"assigned_to"`
	LostReason    *string    `json:"lost_reason"`
	CreatedBy     *uuid.UUID `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Contract struct {
	ID            uuid.UUID  `json:"id"`
	Code          string     `json:"code"`
	CustomerID    uuid.UUID  `json:"customer_id"`
	OpportunityID *uuid.UUID `json:"opportunity_id"`
	Title         *string    `json:"title"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	ServiceType   *string    `json:"service_type"`
	Value         *float64   `json:"value"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	PaymentTerms  *string    `json:"payment_terms"`
	FileURL       *string    `json:"file_url"`
	CreatedBy     *uuid.UUID `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Quotation struct {
	ID         uuid.UUID  `json:"id"`
	Code       string     `json:"code"`
	CustomerID uuid.UUID  `json:"customer_id"`
	OppID      *uuid.UUID `json:"opp_id"`
	Items      []LineItem `json:"items"`
	Total      *float64   `json:"total"`
	Currency   string     `json:"currency"`
	ValidUntil *time.Time `json:"valid_until"`
	Status     string     `json:"status"`
	CreatedBy  *uuid.UUID `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
}

type OpportunityFilter struct {
	Page, Limit, Offset int
	Stage, AssignedTo   string
	UserID, Role        string
}

type OpportunityRepository interface {
	Create(ctx context.Context, o *Opportunity) error
	List(ctx context.Context, f OpportunityFilter) ([]Opportunity, int, error)
	Get(ctx context.Context, id uuid.UUID) (*Opportunity, error)
	Update(ctx context.Context, o *Opportunity) error
	Delete(ctx context.Context, id uuid.UUID) error
	InsertStageHistory(ctx context.Context, oppID uuid.UUID, fromStage, toStage string, changedBy *uuid.UUID) error
}

type ContractRepository interface {
	NextCode(ctx context.Context, prefix, table string) (string, error)
	Create(ctx context.Context, c *Contract) error
	List(ctx context.Context, status, customerID string, limit, offset int) ([]Contract, int, error)
	Get(ctx context.Context, id uuid.UUID) (*Contract, error)
	Update(ctx context.Context, c *Contract) error
}

type QuotationRepository interface {
	NextCode(ctx context.Context, prefix, table string) (string, error)
	Create(ctx context.Context, q *Quotation, itemsJSON []byte) error
	List(ctx context.Context, status, customerID string, limit, offset int) ([]Quotation, int, error)
	Get(ctx context.Context, id uuid.UUID) (*Quotation, error)
	Update(ctx context.Context, q *Quotation, itemsJSON []byte) error
	GetCustomerEmail(ctx context.Context, customerID uuid.UUID) (string, error)
}

type MailerPort interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}
