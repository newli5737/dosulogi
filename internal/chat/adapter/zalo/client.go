package zalo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/chat/domain"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *Client) Available() bool {
	return c.baseURL != ""
}

func (c *Client) CreateAccount(name string) (string, error) {
	var out struct {
		ID string `json:"id"`
	}
	if err := c.post("/api/accounts", map[string]string{"name": name}, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (c *Client) StartQRLogin(bridgeID string) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := c.post("/api/accounts/"+bridgeID+"/login-qr", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) LoginStatus(bridgeID string) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := c.get("/api/accounts/"+bridgeID+"/login-status", &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) FetchInbox(bridgeID string) (*domain.ChatInboxResponse, error) {
	var out domain.ChatInboxResponse
	if err := c.get("/api/accounts/"+bridgeID+"/inbox", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) FetchThread(bridgeID, threadID, cursor string) (*domain.ChatThreadResponse, error) {
	path := fmt.Sprintf("/api/accounts/%s/threads/%s?cursor=%s", bridgeID, threadID, cursor)
	var out domain.ChatThreadResponse
	if err := c.get(path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SendMessage(bridgeID, threadID, text string) error {
	return c.post("/api/accounts/"+bridgeID+"/send", map[string]string{
		"thread_id": threadID,
		"text":      text,
	}, &struct{}{})
}

func (c *Client) get(path string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decode(resp, out)
}

func (c *Client) post(path string, body interface{}, out interface{}) error {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decode(resp, out)
}

func decode(resp *http.Response, out interface{}) error {
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var errBody struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(body, &errBody)
		if errBody.Error != "" {
			return fmt.Errorf("%s", errBody.Error)
		}
		return fmt.Errorf("zalo bridge status %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode zalo bridge response: %w", err)
	}
	return nil
}
