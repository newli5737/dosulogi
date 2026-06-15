package crm

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID            uuid.UUID  `json:"id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	Email         *string    `json:"email"`
	Phone         *string    `json:"phone"`
	Address       *string    `json:"address"`
	Province      *string    `json:"province"`
	Segment       *string    `json:"segment"`
	Tier          *string    `json:"tier"`
	AssignedTo    *uuid.UUID `json:"assigned_to"`
	LastContactAt *time.Time `json:"last_contact_at"`
	CreatedBy     *uuid.UUID `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Contact struct {
	ID         uuid.UUID `json:"id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Name       string    `json:"name"`
	Role       *string   `json:"role"`
	Phone      *string   `json:"phone"`
	Email      *string   `json:"email"`
	IsPrimary  bool      `json:"is_primary"`
	CreatedAt  time.Time `json:"created_at"`
}

type Interaction struct {
	ID         uuid.UUID  `json:"id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	Channel    string     `json:"channel"`
	Direction  *string    `json:"direction"`
	Summary    *string    `json:"summary"`
	CreatedBy  *uuid.UUID `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateCustomerRequest struct {
	Name       string     `json:"name" binding:"required"`
	Type       string     `json:"type" binding:"required,oneof=B2B B2C"`
	Email      *string    `json:"email"`
	Phone      *string    `json:"phone"`
	Address    *string    `json:"address"`
	Province   *string    `json:"province"`
	Segment    *string    `json:"segment"`
	Tier       *string    `json:"tier"`
	AssignedTo *uuid.UUID `json:"assigned_to"`
}

type UpdateCustomerRequest struct {
	Name       *string    `json:"name"`
	Type       *string    `json:"type"`
	Email      *string    `json:"email"`
	Phone      *string    `json:"phone"`
	Address    *string    `json:"address"`
	Province   *string    `json:"province"`
	Segment    *string    `json:"segment"`
	Tier       *string    `json:"tier"`
	AssignedTo *uuid.UUID `json:"assigned_to"`
}

type CreateContactRequest struct {
	Name      string  `json:"name" binding:"required"`
	Role      *string `json:"role"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
	IsPrimary bool    `json:"is_primary"`
}

type CreateInteractionRequest struct {
	Channel   string  `json:"channel" binding:"required"`
	Direction *string `json:"direction"`
	Summary   *string `json:"summary"`
}

type CustomerFilter struct {
	Segment    string
	Tier       string
	AssignedTo string
	Query      string
	UserID     string
	Role       string
	Limit      int
	Offset     int
}

func (c *Customer) MarshalJSON() ([]byte, error) {
	type Alias Customer
	return json.Marshal((*Alias)(c))
}
