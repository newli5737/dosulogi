package tracking3p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type ShipmentData struct {
	TrackingCode string         `json:"tracking_code"`
	Status       string         `json:"status"`
	Origin       string         `json:"origin"`
	Destination  string         `json:"destination"`
	Lat          float64        `json:"lat"`
	Lng          float64        `json:"lng"`
	Events       []ShipmentEvent `json:"events"`
}

type ShipmentEvent struct {
	Status      string     `json:"status"`
	Description string     `json:"description"`
	Location    string     `json:"location"`
	Lat         float64    `json:"lat"`
	Lng         float64    `json:"lng"`
	EventTime   *time.Time `json:"event_time"`
}

func (c *Client) FetchShipment(ctx context.Context, trackingCode string) (*ShipmentData, error) {
	if c.apiKey == "" {
		// stub response for development
		return &ShipmentData{
			TrackingCode: trackingCode,
			Status:       "in_transit",
			Origin:       "Ho Chi Minh",
			Destination:  "Ha Noi",
			Lat:          16.0544,
			Lng:          108.2022,
		}, nil
	}

	url := fmt.Sprintf("%s/v1/shipments/%s", c.baseURL, trackingCode)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("tracking API error: %d", resp.StatusCode)
	}

	var data ShipmentData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *Client) FetchShipments(ctx context.Context, ids []string) ([]ShipmentData, error) {
	var results []ShipmentData
	for _, id := range ids {
		d, err := c.FetchShipment(ctx, id)
		if err != nil {
			continue
		}
		results = append(results, *d)
	}
	return results, nil
}
