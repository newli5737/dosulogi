package marketing

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Campaign struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Status      string          `json:"status"`
	Subject     *string         `json:"subject"`
	BodyHTML    *string         `json:"body_html"`
	Segment     json.RawMessage `json:"segment"`
	ScheduledAt *time.Time      `json:"scheduled_at"`
	SentCount   int             `json:"sent_count"`
	CreatedBy   *uuid.UUID      `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CampaignLog struct {
	ID          uuid.UUID  `json:"id"`
	CampaignID  uuid.UUID  `json:"campaign_id"`
	CustomerID  *uuid.UUID `json:"customer_id"`
	Email       *string    `json:"email"`
	Status      *string    `json:"status"`
	SGMessageID *string    `json:"sg_message_id"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ScheduleRequest struct {
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
}

type EmailWebhookEvent struct {
	Email     string `json:"email"`
	Event     string `json:"event"`
	MessageID string `json:"message_id"`
}
