package marketing

import (
	"context"

	"github.com/dosu-logi/logistics-erp/internal/integration/mailer"
	"github.com/google/uuid"
)

type Service struct {
	repo   *Repository
	mailer *mailer.Client
}

func NewService(repo *Repository, mailer *mailer.Client) *Service {
	return &Service{repo: repo, mailer: mailer}
}

func (s *Service) Create(ctx context.Context, c *Campaign) error {
	if c.Status == "" {
		c.Status = "draft"
	}
	return s.repo.Create(ctx, c)
}

func (s *Service) List(ctx context.Context, status string, limit, offset int) ([]Campaign, int, error) {
	return s.repo.List(ctx, status, limit, offset)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Campaign, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Update(ctx context.Context, c *Campaign) error {
	return s.repo.Update(ctx, c)
}

func (s *Service) Send(ctx context.Context, id uuid.UUID) error {
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	return s.sendCampaign(ctx, c)
}

func (s *Service) sendCampaign(ctx context.Context, c *Campaign) error {
	targets, err := s.repo.ListTargetEmails(ctx, c.Segment)
	if err != nil {
		return err
	}
	subject := ""
	if c.Subject != nil {
		subject = *c.Subject
	}
	body := ""
	if c.BodyHTML != nil {
		body = *c.BodyHTML
	}
	sent := 0
	for _, t := range targets {
		msgID, err := s.mailer.SendBulkEmail(ctx, t.Email, subject, body)
		status := "sent"
		if err != nil {
			status = "bounced"
		} else {
			sent++
		}
		email := t.Email
		cid := t.ID
		log := &CampaignLog{CampaignID: c.ID, CustomerID: &cid, Email: &email, Status: &status}
		if msgID != "" {
			log.SGMessageID = &msgID
		}
		_ = s.repo.CreateLog(ctx, log)
	}
	c.Status = "sending"
	_ = s.repo.Update(ctx, c)
	return s.repo.IncrementSentCount(ctx, c.ID, sent)
}

func (s *Service) ListLogs(ctx context.Context, campaignID uuid.UUID, status string, limit, offset int) ([]CampaignLog, int, error) {
	return s.repo.ListLogs(ctx, campaignID, status, limit, offset)
}

func (s *Service) HandleEmailWebhook(ctx context.Context, events []EmailWebhookEvent) error {
	for _, ev := range events {
		status := mapEmailEvent(ev.Event)
		if status != "" && ev.MessageID != "" {
			_ = s.repo.UpdateLogByMessageID(ctx, ev.MessageID, status)
		}
	}
	return nil
}

func mapEmailEvent(event string) string {
	switch event {
	case "delivered":
		return "delivered"
	case "open":
		return "opened"
	case "click":
		return "clicked"
	case "bounce", "dropped":
		return "bounced"
	default:
		return ""
	}
}

func (s *Service) ProcessScheduled(ctx context.Context) error {
	campaigns, err := s.repo.ListScheduledDue(ctx)
	if err != nil {
		return err
	}
	for i := range campaigns {
		_ = s.sendCampaign(ctx, &campaigns[i])
	}
	return nil
}
