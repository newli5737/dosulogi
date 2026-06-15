package mailer

import (
	"context"
	"log"
)

// Client logs outbound email when no SMTP is configured.
// Replace with SMTP/SES later if needed.
type Client struct {
	fromEmail string
}

func New(fromEmail string) *Client {
	return &Client{fromEmail: fromEmail}
}

func (c *Client) SendEmail(ctx context.Context, to, subject, htmlBody string) error {
	log.Printf("[mailer] to=%s from=%s subject=%s", to, c.fromEmail, subject)
	return nil
}

func (c *Client) SendEmailWithAttachment(ctx context.Context, to, subject, htmlBody string, filePath *string) error {
	path := ""
	if filePath != nil {
		path = *filePath
	}
	log.Printf("[mailer] to=%s subject=%s attachment=%s", to, subject, path)
	return nil
}

func (c *Client) SendBulkEmail(ctx context.Context, to, subject, htmlBody string) (string, error) {
	log.Printf("[mailer] bulk to=%s subject=%s", to, subject)
	return "local-" + to, nil
}
