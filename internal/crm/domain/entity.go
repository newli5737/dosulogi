package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserBrief struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
}

type Customer struct {
	ID            uuid.UUID  `json:"id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	Email         *string    `json:"email"`
	Phone         *string    `json:"phone"`
	Address       *string    `json:"address"`
	Province      *string    `json:"province"`
	TaxCode       *string    `json:"tax_code"`
	Segment       string     `json:"segment"`
	Tier          string     `json:"tier"`
	AssignedTo    *uuid.UUID `json:"-"`
	AssignedUser  *UserBrief `json:"assigned_to"`
	LastContactAt *time.Time `json:"last_contact_at"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CustomerFilter struct {
	Page, Limit, Offset int
	Query, Type, Segment, Tier, Province string
	AssignedTo string
	IsActive   *bool
	UserID, Role string
}

type CustomerRepository interface {
	List(ctx context.Context, f CustomerFilter) ([]Customer, int, error)
	Get(ctx context.Context, id uuid.UUID) (*Customer, error)
	Create(ctx context.Context, c *Customer) error
	Update(ctx context.Context, c *Customer) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	EmailExists(ctx context.Context, email string, excludeID *uuid.UUID) (bool, error)
	NextCode(ctx context.Context) (string, error)
	CountOpenTickets(ctx context.Context, customerID uuid.UUID) (int, error)
	CountActiveContracts(ctx context.Context, customerID uuid.UUID) (int, error)
}

type CustomerDetail struct {
	Customer
	PrimaryContact  *Contact `json:"primary_contact"`
	OpenTickets     int      `json:"open_tickets"`
	ActiveContracts int      `json:"active_contracts"`
}

type Interaction struct {
	ID         uuid.UUID  `json:"id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	Channel    string     `json:"channel"`
	Direction  *string    `json:"direction"`
	Summary    string     `json:"summary"`
	OccurredAt time.Time  `json:"occurred_at"`
	CreatedBy  *UserBrief `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
}

type InteractionFilter struct {
	Page, Limit, Offset int
	CustomerID          uuid.UUID
	Channel             string
}

type InteractionRepository interface {
	List(ctx context.Context, f InteractionFilter) ([]Interaction, int, error)
	Create(ctx context.Context, i *Interaction, createdBy uuid.UUID) error
	TouchLastContact(ctx context.Context, customerID uuid.UUID, at time.Time) error
}

type Contact struct {
	ID         uuid.UUID `json:"id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Name       string    `json:"name"`
	Role       *string   `json:"role"`
	Phone      *string   `json:"phone"`
	Email      *string   `json:"email"`
	IsPrimary  bool      `json:"is_primary"`
	Note       *string   `json:"note"`
}

type ContactRepository interface {
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]Contact, error)
	Create(ctx context.Context, c *Contact) error
	Update(ctx context.Context, c *Contact) error
	Delete(ctx context.Context, customerID, contactID uuid.UUID) error
	UnsetPrimary(ctx context.Context, customerID uuid.UUID) error
	CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error)
}

type Ticket struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`
	CustomerID  uuid.UUID  `json:"customer_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	Category    *string    `json:"category"`
	AssignedTo  *uuid.UUID `json:"-"`
	SLADeadline *time.Time `json:"sla_deadline"`
	IsOverdue   bool       `json:"is_overdue"`
	CreatedAt   time.Time  `json:"created_at"`
	Customer    *struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
		Code string    `json:"code"`
	} `json:"customer,omitempty"`
	AssignedUser *UserBrief `json:"assigned_to"`
}

type TicketFilter struct {
	Page, Limit, Offset int
	Status, Priority, AssignedTo, CustomerID string
	Overdue bool
}

type TicketRepository interface {
	List(ctx context.Context, f TicketFilter) ([]Ticket, int, error)
	Get(ctx context.Context, id uuid.UUID) (*Ticket, error)
	Create(ctx context.Context, t *Ticket) error
	Update(ctx context.Context, t *Ticket) error
	NextCode(ctx context.Context) (string, error)
	GetCustomerAssignee(ctx context.Context, customerID uuid.UUID) (*uuid.UUID, error)
}

type TicketComment struct {
	ID         uuid.UUID  `json:"id"`
	TicketID   uuid.UUID  `json:"ticket_id"`
	Body       string     `json:"body"`
	IsInternal bool       `json:"is_internal"`
	CreatedBy  *UserBrief `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
}

type TicketCommentRepository interface {
	ListByTicket(ctx context.Context, ticketID uuid.UUID) ([]TicketComment, error)
	Create(ctx context.Context, c *TicketComment, createdBy uuid.UUID) error
}
