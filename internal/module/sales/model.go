package sales

import (
	"encoding/json"
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

func ItemsToJSON(items []LineItem) ([]byte, error) {
	return json.Marshal(items)
}

func ItemsFromJSON(data []byte) ([]LineItem, error) {
	var items []LineItem
	err := json.Unmarshal(data, &items)
	return items, err
}
