package sepay

import (
	"encoding/json"
	"io"
	"strings"
)

type WebhookPayload struct {
	ID              string  `json:"id"`
	Gateway         string  `json:"gateway"`
	TransactionDate string  `json:"transactionDate"`
	AccountNumber   string  `json:"accountNumber"`
	Code            string  `json:"code"`
	Content         string  `json:"content"`
	TransferType    string  `json:"transferType"`
	TransferAmount  float64 `json:"transferAmount"`
	Accumulated     float64 `json:"accumulated"`
	ReferenceCode   string  `json:"referenceCode"`
	Description     string  `json:"description"`
}

func VerifyAuth(header, secret string) bool {
	expected := "Apikey " + secret
	return strings.TrimSpace(header) == expected
}

func ParseWebhook(r io.Reader) (WebhookPayload, error) {
	var p WebhookPayload
	err := json.NewDecoder(r).Decode(&p)
	if p.Code == "" && p.Content != "" {
		// try extract invoice code from content
		p.Code = extractInvoiceCode(p.Content)
	}
	return p, err
}

func extractInvoiceCode(content string) string {
	// look for INV-YYYY-NNN pattern
	parts := strings.Fields(content)
	for _, p := range parts {
		if strings.HasPrefix(p, "INV-") {
			return p
		}
	}
	return content
}
