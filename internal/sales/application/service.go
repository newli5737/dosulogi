package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/sales/domain"
	"github.com/google/uuid"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation")
)

type OpportunityService struct {
	repo domain.OpportunityRepository
}

func NewOpportunityService(r domain.OpportunityRepository) *OpportunityService {
	return &OpportunityService{repo: r}
}

func (s *OpportunityService) List(ctx context.Context, f domain.OpportunityFilter) ([]domain.Opportunity, int, error) {
	return s.repo.List(ctx, f)
}

func (s *OpportunityService) Get(ctx context.Context, id uuid.UUID) (*domain.Opportunity, error) {
	return s.repo.Get(ctx, id)
}

func (s *OpportunityService) ListStageHistory(ctx context.Context, id uuid.UUID) ([]domain.StageHistoryEntry, error) {
	return s.repo.ListStageHistory(ctx, id)
}

type CreateOpportunityInput struct {
	domain.Opportunity
	ShipmentIDs []uuid.UUID `json:"shipment_ids"`
}

func (s *OpportunityService) Create(ctx context.Context, in *CreateOpportunityInput) error {
	o := &in.Opportunity
	if o.Title == "" || o.CustomerID == uuid.Nil {
		return ErrValidation
	}
	code, err := s.repo.NextCode(ctx)
	if err != nil {
		return err
	}
	o.Code = code
	if o.Stage == "" {
		o.Stage = "lead"
	}
	if o.Currency == "" {
		o.Currency = "VND"
	}
	if o.AssignedTo == nil {
		o.AssignedTo = o.CreatedBy
	}
	if err := s.repo.Create(ctx, o); err != nil {
		return err
	}
	if len(in.ShipmentIDs) > 0 {
		return s.repo.SetShipments(ctx, o.ID, in.ShipmentIDs)
	}
	return nil
}

func (s *OpportunityService) Update(ctx context.Context, o *domain.Opportunity, shipmentIDs []uuid.UUID, changedBy uuid.UUID) error {
	existing, err := s.repo.Get(ctx, o.ID)
	if err != nil {
		return err
	}
	if existing.Stage != o.Stage {
		cb := changedBy
		if err := s.repo.InsertStageHistory(ctx, o.ID, existing.Stage, o.Stage, &cb); err != nil {
			return err
		}
	}
	if err := s.repo.Update(ctx, o); err != nil {
		return err
	}
	if shipmentIDs != nil {
		return s.repo.SetShipments(ctx, o.ID, shipmentIDs)
	}
	return nil
}

func (s *OpportunityService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

type QuotationService struct {
	quotes   domain.QuotationRepository
	contracts domain.ContractRepository
	mailer   domain.MailerPort
}

func NewQuotationService(q domain.QuotationRepository, c domain.ContractRepository, m domain.MailerPort) *QuotationService {
	return &QuotationService{quotes: q, contracts: c, mailer: m}
}

func calcTotal(items []domain.LineItem) float64 {
	var total float64
	for _, it := range items {
		amt := it.Qty * it.UnitPrice
		if it.Amount > 0 {
			amt = it.Amount
		}
		total += amt
	}
	return total
}

func itemsToJSON(items []domain.LineItem) ([]byte, error) {
	return json.Marshal(items)
}

func (s *QuotationService) Create(ctx context.Context, qt *domain.Quotation) error {
	code, err := s.quotes.NextCode(ctx, "BG", "quotations")
	if err != nil {
		return err
	}
	qt.Code = code
	total := calcTotal(qt.Items)
	qt.Total = &total
	if qt.Status == "" {
		qt.Status = "draft"
	}
	itemsJSON, _ := itemsToJSON(qt.Items)
	return s.quotes.Create(ctx, qt, itemsJSON)
}

func (s *QuotationService) List(ctx context.Context, status, customerID string, limit, offset int) ([]domain.Quotation, int, error) {
	return s.quotes.List(ctx, status, customerID, limit, offset)
}

func (s *QuotationService) Get(ctx context.Context, id uuid.UUID) (*domain.Quotation, error) {
	return s.quotes.Get(ctx, id)
}

func (s *QuotationService) Update(ctx context.Context, qt *domain.Quotation) error {
	total := calcTotal(qt.Items)
	qt.Total = &total
	itemsJSON, _ := itemsToJSON(qt.Items)
	return s.quotes.Update(ctx, qt, itemsJSON)
}

func (s *QuotationService) Send(ctx context.Context, id uuid.UUID) error {
	qt, err := s.quotes.Get(ctx, id)
	if err != nil {
		return err
	}
	email, err := s.quotes.GetCustomerEmail(ctx, qt.CustomerID)
	if err != nil || email == "" {
		return err
	}
	if err := s.mailer.SendEmail(ctx, email, "Báo giá "+qt.Code, "<p>Báo giá "+qt.Code+"</p>"); err != nil {
		return err
	}
	qt.Status = "sent"
	itemsJSON, _ := itemsToJSON(qt.Items)
	return s.quotes.Update(ctx, qt, itemsJSON)
}

func (s *QuotationService) Convert(ctx context.Context, id uuid.UUID, createdBy uuid.UUID) (*domain.Contract, error) {
	qt, err := s.quotes.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	title := "Contract from " + qt.Code
	now := time.Now()
	c := &domain.Contract{
		CustomerID: qt.CustomerID, OpportunityID: qt.OppID,
		Title: &title, StartDate: now, Value: qt.Total, Currency: qt.Currency,
		Status: "draft", CreatedBy: &createdBy,
	}
	svc := &ContractService{repo: s.contracts}
	if err := svc.Create(ctx, c); err != nil {
		return nil, err
	}
	qt.Status = "accepted"
	itemsJSON, _ := itemsToJSON(qt.Items)
	_ = s.quotes.Update(ctx, qt, itemsJSON)
	return c, nil
}

type ContractService struct {
	repo domain.ContractRepository
}

func NewContractService(r domain.ContractRepository) *ContractService {
	return &ContractService{repo: r}
}

func (s *ContractService) Create(ctx context.Context, c *domain.Contract) error {
	code, err := s.repo.NextCode(ctx, "HD", "contracts")
	if err != nil {
		return err
	}
	c.Code = code
	if c.Status == "" {
		c.Status = "draft"
	}
	return s.repo.Create(ctx, c)
}

func (s *ContractService) List(ctx context.Context, status, customerID string, limit, offset int) ([]domain.Contract, int, error) {
	return s.repo.List(ctx, status, customerID, limit, offset)
}

func (s *ContractService) Get(ctx context.Context, id uuid.UUID) (*domain.Contract, error) {
	return s.repo.Get(ctx, id)
}

func (s *ContractService) Update(ctx context.Context, c *domain.Contract) error {
	return s.repo.Update(ctx, c)
}

func (s *ContractService) UploadFile(ctx context.Context, id uuid.UUID, fileURL string) (*domain.Contract, error) {
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	c.FileURL = &fileURL
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}
