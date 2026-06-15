package mailer

import (
	"context"

	"github.com/dosu-logi/logistics-erp/internal/integration/mailer"
)

type Adapter struct{ client *mailer.Client }

func New(client *mailer.Client) *Adapter { return &Adapter{client: client} }

func (a *Adapter) SendEmail(ctx context.Context, to, subject, body string) error {
	return a.client.SendEmail(ctx, to, subject, body)
}
