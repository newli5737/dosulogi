package fbchat

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dosu-logi/logistics-erp/internal/chat/domain"
)

// VerifySession checks whether Facebook cookies can load the home page and extract fb_dtsg.
func VerifySession(cookiesJSON string) (facebookID string, err error) {
	cookieHeader, err := BuildCookieHeader(cookiesJSON)
	if err != nil {
		return "", err
	}
	tokens, err := fetchFBTokens(cookieHeader)
	if err != nil {
		return "", err
	}
	if tokens.FacebookID != "" {
		return tokens.FacebookID, nil
	}
	return "", fmt.Errorf("facebook session ok but actor id not found")
}

// BuildCookieHeader converts stored cookie JSON or raw string to a Cookie header value.
func BuildCookieHeader(cookiesJSON string) (string, error) {
	trimmed := strings.TrimSpace(cookiesJSON)
	if trimmed == "" {
		return "", fmt.Errorf("empty cookies")
	}
	if strings.Contains(trimmed, "=") && !strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "{") {
		return trimmed, nil
	}

	var export domain.CookieExport
	if err := json.Unmarshal([]byte(trimmed), &export); err != nil {
		var cookies []domain.Cookie
		if err2 := json.Unmarshal([]byte(trimmed), &cookies); err2 != nil {
			return "", fmt.Errorf("invalid cookies format")
		}
		export.Cookies = cookies
	}

	var parts []string
	for _, c := range export.Cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("no cookies found")
	}
	return strings.Join(parts, "; "), nil
}
