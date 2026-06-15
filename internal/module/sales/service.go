package sales

import (
	"context"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/integration/mailer"
	"github.com/google/uuid"
)

type Service struct {
	repo     *Repository
	mailer   *mailer.Client
	uploadDir string
}

func NewService(repo *Repository, mailer *mailer.Client, uploadDir string) *Service {
	return &Service{repo: repo, mailer: mailer, uploadDir: uploadDir}
}

func calcTotal(items []LineItem) float64 {
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

func (s *Service) CreateOpportunity(ctx context.Context, o *Opportunity) error {
	if o.AssignedTo == nil {
		o.AssignedTo = o.CreatedBy
	}
	return s.repo.CreateOpportunity(ctx, o)
}

func (s *Service) ListOpportunities(ctx context.Context, stage, assignedTo, userID, role string, limit, offset int) ([]Opportunity, int, error) {
	return s.repo.ListOpportunities(ctx, stage, assignedTo, userID, role, limit, offset)
}

func (s *Service) GetOpportunity(ctx context.Context, id uuid.UUID) (*Opportunity, error) {
	return s.repo.GetOpportunity(ctx, id)
}

func (s *Service) UpdateOpportunity(ctx context.Context, o *Opportunity, changedBy uuid.UUID) error {
	existing, err := s.repo.GetOpportunity(ctx, o.ID)
	if err != nil {
		return err
	}
	if existing.Stage != o.Stage {
		cb := changedBy
		if err := s.repo.InsertStageHistory(ctx, o.ID, existing.Stage, o.Stage, &cb); err != nil {
			return err
		}
	}
	return s.repo.UpdateOpportunity(ctx, o)
}

func (s *Service) DeleteOpportunity(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteOpportunity(ctx, id)
}

func (s *Service) CreateContract(ctx context.Context, c *Contract) error {
	code, err := s.repo.NextCode(ctx, "HD", "contracts")
	if err != nil {
		return err
	}
	c.Code = code
	if c.Status == "" {
		c.Status = "draft"
	}
	return s.repo.CreateContract(ctx, c)
}

func (s *Service) ListContracts(ctx context.Context, status, customerID string, limit, offset int) ([]Contract, int, error) {
	return s.repo.ListContracts(ctx, status, customerID, limit, offset)
}

func (s *Service) GetContract(ctx context.Context, id uuid.UUID) (*Contract, error) {
	return s.repo.GetContract(ctx, id)
}

func (s *Service) UpdateContract(ctx context.Context, c *Contract) error {
	return s.repo.UpdateContract(ctx, c)
}

func (s *Service) UploadContractFile(ctx context.Context, id uuid.UUID, fileURL string) (*Contract, error) {
	c, err := s.repo.GetContract(ctx, id)
	if err != nil {
		return nil, err
	}
	c.FileURL = &fileURL
	if err := s.repo.UpdateContract(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) CreateQuotation(ctx context.Context, qt *Quotation) error {
	code, err := s.repo.NextCode(ctx, "BG", "quotations")
	if err != nil {
		return err
	}
	qt.Code = code
	total := calcTotal(qt.Items)
	qt.Total = &total
	if qt.Status == "" {
		qt.Status = "draft"
	}
	itemsJSON, _ := ItemsToJSON(qt.Items)
	return s.repo.CreateQuotation(ctx, qt, itemsJSON)
}

func (s *Service) ListQuotations(ctx context.Context, status, customerID string, limit, offset int) ([]Quotation, int, error) {
	return s.repo.ListQuotations(ctx, status, customerID, limit, offset)
}

func (s *Service) GetQuotation(ctx context.Context, id uuid.UUID) (*Quotation, error) {
	return s.repo.GetQuotation(ctx, id)
}

func (s *Service) UpdateQuotation(ctx context.Context, qt *Quotation) error {
	total := calcTotal(qt.Items)
	qt.Total = &total
	itemsJSON, _ := ItemsToJSON(qt.Items)
	return s.repo.UpdateQuotation(ctx, qt, itemsJSON)
}

func (s *Service) SendQuotation(ctx context.Context, id uuid.UUID) error {
	qt, err := s.repo.GetQuotation(ctx, id)
	if err != nil {
		return err
	}
	email, err := s.repo.GetCustomerEmail(ctx, qt.CustomerID)
	if err != nil || email == "" {
		return err
	}
	subject := "Báo giá " + qt.Code
	body := "<p>Kính gửi quý khách,</p><p>Chúng tôi gửi kèm báo giá " + qt.Code + ".</p>"
	if err := s.mailer.SendEmail(ctx, email, subject, body); err != nil {
		return err
	}
	qt.Status = "sent"
	itemsJSON, _ := ItemsToJSON(qt.Items)
	return s.repo.UpdateQuotation(ctx, qt, itemsJSON)
}

func (s *Service) ConvertQuotation(ctx context.Context, id uuid.UUID, createdBy uuid.UUID) (*Contract, error) {
	qt, err := s.repo.GetQuotation(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	c := &Contract{
		CustomerID: qt.CustomerID, OpportunityID: qt.OppID,
		Title: strPtr("Contract from " + qt.Code),
		StartDate: now, Value: qt.Total, Currency: qt.Currency,
		Status: "draft", CreatedBy: &createdBy,
	}
	if err := s.CreateContract(ctx, c); err != nil {
		return nil, err
	}
	qt.Status = "accepted"
	itemsJSON, _ := ItemsToJSON(qt.Items)
	_ = s.repo.UpdateQuotation(ctx, qt, itemsJSON)
	return c, nil
}

func strPtr(s string) *string { return &s }
