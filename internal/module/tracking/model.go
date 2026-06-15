package tracking

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Shipment struct {
	ID                uuid.UUID       `json:"id"`
	TrackingCode      string          `json:"tracking_code"`
	ExternalID        *string         `json:"external_id"`
	CustomerID        *uuid.UUID      `json:"customer_id"`
	ContractID        *uuid.UUID      `json:"contract_id"`
	Status            *string         `json:"status"`
	Origin            *string         `json:"origin"`
	Destination       *string         `json:"destination"`
	Lat               *float64        `json:"lat"`
	Lng               *float64        `json:"lng"`
	EstimatedDelivery *time.Time      `json:"estimated_delivery"`
	ActualDelivery    *time.Time      `json:"actual_delivery"`
	RawPayload        json.RawMessage `json:"raw_payload,omitempty"`
	LastSyncedAt      *time.Time      `json:"last_synced_at"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type ShipmentEvent struct {
	ID          uuid.UUID  `json:"id"`
	ShipmentID  uuid.UUID  `json:"shipment_id"`
	Status      *string    `json:"status"`
	Description *string    `json:"description"`
	Location    *string    `json:"location"`
	Lat         *float64   `json:"lat"`
	Lng         *float64   `json:"lng"`
	EventTime   *time.Time `json:"event_time"`
	CreatedAt   time.Time  `json:"created_at"`
}

type WebhookPayload struct {
	TrackingCode string     `json:"tracking_code"`
	Status       string     `json:"status"`
	Location     string     `json:"location"`
	Lat          float64    `json:"lat"`
	Lng          float64    `json:"lng"`
	EventTime    *time.Time `json:"event_time"`
	Description  string     `json:"description"`
}

type MapPoint struct {
	TrackingCode  string  `json:"tracking_code"`
	Status        string  `json:"status"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	CustomerName  string  `json:"customer_name"`
	Destination   string  `json:"destination"`
}
