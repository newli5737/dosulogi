package crm

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCustomer(ctx context.Context, req CreateCustomerRequest, createdBy uuid.UUID) (*Customer, error) {
	code, err := s.repo.NextCustomerCode(ctx)
	if err != nil {
		return nil, err
	}
	c := &Customer{
		Code: code, Name: req.Name, Type: req.Type,
		Email: req.Email, Phone: req.Phone, Address: req.Address,
		Province: req.Province, Segment: req.Segment, Tier: req.Tier,
		AssignedTo: req.AssignedTo, CreatedBy: &createdBy,
	}
	if c.AssignedTo == nil {
		c.AssignedTo = &createdBy
	}
	if err := s.repo.CreateCustomer(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) ListCustomers(ctx context.Context, f CustomerFilter) ([]Customer, int, error) {
	return s.repo.ListCustomers(ctx, f)
}

func (s *Service) GetCustomer(ctx context.Context, id uuid.UUID) (*Customer, error) {
	return s.repo.GetCustomer(ctx, id)
}

func (s *Service) UpdateCustomer(ctx context.Context, id uuid.UUID, req UpdateCustomerRequest, userID, role string) (*Customer, error) {
	c, err := s.repo.GetCustomer(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == "sales_rep" && (c.AssignedTo == nil || c.AssignedTo.String() != userID) {
		return nil, ErrNotFound
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Type != nil {
		c.Type = *req.Type
	}
	if req.Email != nil {
		c.Email = req.Email
	}
	if req.Phone != nil {
		c.Phone = req.Phone
	}
	if req.Address != nil {
		c.Address = req.Address
	}
	if req.Province != nil {
		c.Province = req.Province
	}
	if req.Segment != nil {
		c.Segment = req.Segment
	}
	if req.Tier != nil {
		c.Tier = req.Tier
	}
	if req.AssignedTo != nil {
		c.AssignedTo = req.AssignedTo
	}
	if err := s.repo.UpdateCustomer(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteCustomer(ctx, id)
}

func (s *Service) ListContacts(ctx context.Context, customerID uuid.UUID) ([]Contact, error) {
	return s.repo.ListContacts(ctx, customerID)
}

func (s *Service) CreateContact(ctx context.Context, customerID uuid.UUID, req CreateContactRequest) (*Contact, error) {
	ct := &Contact{CustomerID: customerID, Name: req.Name, Role: req.Role, Phone: req.Phone, Email: req.Email, IsPrimary: req.IsPrimary}
	if err := s.repo.CreateContact(ctx, ct); err != nil {
		return nil, err
	}
	return ct, nil
}

func (s *Service) UpdateContact(ctx context.Context, customerID, contactID uuid.UUID, req CreateContactRequest) (*Contact, error) {
	ct := &Contact{ID: contactID, CustomerID: customerID, Name: req.Name, Role: req.Role, Phone: req.Phone, Email: req.Email, IsPrimary: req.IsPrimary}
	if err := s.repo.UpdateContact(ctx, ct); err != nil {
		return nil, err
	}
	return ct, nil
}

func (s *Service) DeleteContact(ctx context.Context, customerID, contactID uuid.UUID) error {
	return s.repo.DeleteContact(ctx, customerID, contactID)
}

func (s *Service) ListInteractions(ctx context.Context, customerID uuid.UUID) ([]Interaction, error) {
	return s.repo.ListInteractions(ctx, customerID)
}

func (s *Service) CreateInteraction(ctx context.Context, customerID uuid.UUID, req CreateInteractionRequest, createdBy uuid.UUID) (*Interaction, error) {
	i := &Interaction{CustomerID: customerID, Channel: req.Channel, Direction: req.Direction, Summary: req.Summary, CreatedBy: &createdBy}
	if err := s.repo.CreateInteraction(ctx, i); err != nil {
		return nil, err
	}
	_ = s.repo.TouchLastContact(ctx, customerID)
	return i, nil
}
