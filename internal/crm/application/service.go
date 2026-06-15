package application

import (
	"context"
	"errors"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/crm/domain"
	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrConflict      = errors.New("conflict")
	ErrForbidden     = errors.New("forbidden")
	ErrValidation    = errors.New("validation")
	ErrPrimaryContact = errors.New("primary contact")
)

type CustomerService struct {
	customers    domain.CustomerRepository
	contacts     domain.ContactRepository
	interactions domain.InteractionRepository
}

func NewCustomerService(c domain.CustomerRepository, ct domain.ContactRepository, i domain.InteractionRepository) *CustomerService {
	return &CustomerService{customers: c, contacts: ct, interactions: i}
}

type CreateCustomerInput struct {
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Segment    string     `json:"segment"`
	Tier       string     `json:"tier"`
	Email      *string    `json:"email"`
	Phone      *string    `json:"phone"`
	Address    *string    `json:"address"`
	Province   *string    `json:"province"`
	TaxCode    *string    `json:"tax_code"`
	AssignedTo *uuid.UUID `json:"assigned_to"`
	CreatedBy  uuid.UUID  `json:"-"`
}

func (s *CustomerService) List(ctx context.Context, f domain.CustomerFilter) ([]domain.Customer, int, error) {
	return s.customers.List(ctx, f)
}

func (s *CustomerService) Get(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	return s.customers.Get(ctx, id)
}

func (s *CustomerService) GetDetail(ctx context.Context, id uuid.UUID) (*domain.CustomerDetail, error) {
	c, err := s.customers.Get(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	detail := &domain.CustomerDetail{Customer: *c}
	contacts, _ := s.contacts.ListByCustomer(ctx, id)
	for _, ct := range contacts {
		if ct.IsPrimary {
			cp := ct
			detail.PrimaryContact = &cp
			break
		}
	}
	detail.OpenTickets, _ = s.customers.CountOpenTickets(ctx, id)
	detail.ActiveContracts, _ = s.customers.CountActiveContracts(ctx, id)
	return detail, nil
}

func (s *CustomerService) ListInteractions(ctx context.Context, f domain.InteractionFilter) ([]domain.Interaction, int, error) {
	return s.interactions.List(ctx, f)
}

type CreateInteractionInput struct {
	Channel    string     `json:"channel"`
	Direction  *string    `json:"direction"`
	Summary    string     `json:"summary"`
	OccurredAt *time.Time `json:"occurred_at"`
}

func (s *CustomerService) CreateInteraction(ctx context.Context, customerID uuid.UUID, in CreateInteractionInput, userID uuid.UUID) (*domain.Interaction, error) {
	if in.Channel == "" || in.Summary == "" {
		return nil, ErrValidation
	}
	it := &domain.Interaction{CustomerID: customerID, Channel: in.Channel, Direction: in.Direction, Summary: in.Summary}
	if in.OccurredAt != nil {
		it.OccurredAt = *in.OccurredAt
	}
	if err := s.interactions.Create(ctx, it, userID); err != nil {
		return nil, err
	}
	_ = s.interactions.TouchLastContact(ctx, customerID, it.OccurredAt)
	return it, nil
}

func (s *CustomerService) Create(ctx context.Context, in CreateCustomerInput) (*domain.Customer, error) {
	if in.Name == "" || in.Type == "" {
		return nil, ErrValidation
	}
	if in.Email != nil && *in.Email != "" {
		exists, err := s.customers.EmailExists(ctx, *in.Email, nil)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrConflict
		}
	}
	code, err := s.customers.NextCode(ctx)
	if err != nil {
		return nil, err
	}
	seg, tier := in.Segment, in.Tier
	if seg == "" {
		seg = "standard"
	}
	if tier == "" {
		tier = "standard"
	}
	c := &domain.Customer{
		Code: code, Name: in.Name, Type: in.Type,
		Email: in.Email, Phone: in.Phone, Address: in.Address,
		Province: in.Province, TaxCode: in.TaxCode,
		Segment: seg, Tier: tier, AssignedTo: in.AssignedTo,
		IsActive: true,
	}
	if c.AssignedTo == nil {
		c.AssignedTo = &in.CreatedBy
	}
	if err := s.customers.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) Update(ctx context.Context, id uuid.UUID, patch domain.Customer) (*domain.Customer, error) {
	c, err := s.customers.Get(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	if patch.Name != "" {
		c.Name = patch.Name
	}
	if patch.Type != "" {
		c.Type = patch.Type
	}
	if patch.Email != nil {
		c.Email = patch.Email
	}
	if patch.Phone != nil {
		c.Phone = patch.Phone
	}
	if patch.Address != nil {
		c.Address = patch.Address
	}
	if patch.Province != nil {
		c.Province = patch.Province
	}
	if patch.TaxCode != nil {
		c.TaxCode = patch.TaxCode
	}
	if patch.Segment != "" {
		c.Segment = patch.Segment
	}
	if patch.Tier != "" {
		c.Tier = patch.Tier
	}
	if patch.AssignedTo != nil {
		c.AssignedTo = patch.AssignedTo
	}
	if err := s.customers.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) Delete(ctx context.Context, id uuid.UUID, role string) error {
	if role != "admin" && role != "sales_manager" {
		return ErrForbidden
	}
	return s.customers.SoftDelete(ctx, id)
}

func (s *CustomerService) ListContacts(ctx context.Context, customerID uuid.UUID) ([]domain.Contact, error) {
	return s.contacts.ListByCustomer(ctx, customerID)
}

func (s *CustomerService) CreateContact(ctx context.Context, ct *domain.Contact) (*domain.Contact, error) {
	if ct.IsPrimary {
		_ = s.contacts.UnsetPrimary(ctx, ct.CustomerID)
	}
	if err := s.contacts.Create(ctx, ct); err != nil {
		return nil, err
	}
	return ct, nil
}

func (s *CustomerService) UpdateContact(ctx context.Context, ct *domain.Contact) (*domain.Contact, error) {
	if ct.IsPrimary {
		_ = s.contacts.UnsetPrimary(ctx, ct.CustomerID)
	}
	if err := s.contacts.Update(ctx, ct); err != nil {
		return nil, err
	}
	return ct, nil
}

func (s *CustomerService) DeleteContact(ctx context.Context, customerID, contactID uuid.UUID) error {
	contacts, err := s.contacts.ListByCustomer(ctx, customerID)
	if err != nil {
		return err
	}
	var target *domain.Contact
	for _, c := range contacts {
		if c.ID == contactID {
			target = &c
			break
		}
	}
	if target == nil {
		return ErrNotFound
	}
	if target.IsPrimary && len(contacts) == 1 {
		return ErrPrimaryContact
	}
	return s.contacts.Delete(ctx, customerID, contactID)
}

type TicketService struct {
	tickets  domain.TicketRepository
	comments domain.TicketCommentRepository
}

func NewTicketService(t domain.TicketRepository, c domain.TicketCommentRepository) *TicketService {
	return &TicketService{tickets: t, comments: c}
}

func slaDeadline(priority string, from time.Time) time.Time {
	switch priority {
	case "urgent":
		return from.Add(4 * time.Hour)
	case "high":
		return from.Add(8 * time.Hour)
	case "low":
		return from.Add(72 * time.Hour)
	default:
		return from.Add(24 * time.Hour)
	}
}

type CreateTicketInput struct {
	CustomerID  uuid.UUID `json:"customer_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Priority    string    `json:"priority"`
	Category    *string   `json:"category"`
	CreatedBy   uuid.UUID `json:"-"`
}

func (s *TicketService) List(ctx context.Context, f domain.TicketFilter) ([]domain.Ticket, int, error) {
	return s.tickets.List(ctx, f)
}

func (s *TicketService) Get(ctx context.Context, id uuid.UUID) (*domain.Ticket, []domain.TicketComment, error) {
	t, err := s.tickets.Get(ctx, id)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	comments, _ := s.comments.ListByTicket(ctx, id)
	return t, comments, nil
}

func (s *TicketService) Create(ctx context.Context, in CreateTicketInput) (*domain.Ticket, error) {
	if in.Title == "" {
		return nil, ErrValidation
	}
	priority := in.Priority
	if priority == "" {
		priority = "medium"
	}
	code, err := s.tickets.NextCode(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	deadline := slaDeadline(priority, now)
	assignee, _ := s.tickets.GetCustomerAssignee(ctx, in.CustomerID)
	t := &domain.Ticket{
		Code: code, CustomerID: in.CustomerID, Title: in.Title,
		Description: in.Description, Priority: priority, Category: in.Category,
		Status: "open", AssignedTo: assignee, SLADeadline: &deadline,
	}
	if err := s.tickets.Create(ctx, t); err != nil {
		return nil, err
	}
	_ = s.comments.Create(ctx, &domain.TicketComment{
		TicketID: t.ID, Body: "Ticket được tạo tự động từ hệ thống", IsInternal: true,
	}, in.CreatedBy)
	return t, nil
}

func (s *TicketService) Update(ctx context.Context, id uuid.UUID, patch domain.Ticket) (*domain.Ticket, error) {
	t, err := s.tickets.Get(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	if patch.Status != "" {
		t.Status = patch.Status
	}
	if patch.Priority != "" && patch.Priority != t.Priority {
		t.Priority = patch.Priority
		d := slaDeadline(t.Priority, time.Now())
		t.SLADeadline = &d
	}
	if patch.AssignedTo != nil {
		t.AssignedTo = patch.AssignedTo
	}
	if err := s.tickets.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TicketService) AddComment(ctx context.Context, ticketID uuid.UUID, body string, internal bool, userID uuid.UUID) (*domain.TicketComment, error) {
	c := &domain.TicketComment{TicketID: ticketID, Body: body, IsInternal: internal}
	if err := s.comments.Create(ctx, c, userID); err != nil {
		return nil, err
	}
	return c, nil
}
