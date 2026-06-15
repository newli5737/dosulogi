package tracking

import (
	"context"
	"encoding/json"

	"github.com/dosu-logi/logistics-erp/internal/integration/tracking3p"
	"github.com/google/uuid"
)

type Service struct {
	repo   *Repository
	client *tracking3p.Client
}

func NewService(repo *Repository, client *tracking3p.Client) *Service {
	return &Service{repo: repo, client: client}
}

func (s *Service) List(ctx context.Context, status, customerID string, limit, offset int) ([]Shipment, int, error) {
	return s.repo.List(ctx, status, customerID, limit, offset)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Shipment, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) ListEvents(ctx context.Context, shipmentID uuid.UUID) ([]ShipmentEvent, error) {
	return s.repo.ListEvents(ctx, shipmentID)
}

func (s *Service) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]Shipment, error) {
	return s.repo.ListByCustomer(ctx, customerID)
}

func (s *Service) SyncShipment(ctx context.Context, id uuid.UUID) (*Shipment, error) {
	sh, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.syncOne(ctx, sh)
}

func (s *Service) syncOne(ctx context.Context, sh *Shipment) (*Shipment, error) {
	data, err := s.client.FetchShipment(ctx, sh.TrackingCode)
	if err != nil {
		return nil, err
	}
	if data.Status != "" {
		sh.Status = &data.Status
	}
	if data.Origin != "" {
		sh.Origin = &data.Origin
	}
	if data.Destination != "" {
		sh.Destination = &data.Destination
	}
	if data.Lat != 0 {
		sh.Lat = &data.Lat
	}
	if data.Lng != 0 {
		sh.Lng = &data.Lng
	}
	raw, _ := json.Marshal(data)
	sh.RawPayload = raw
	if err := s.repo.UpdateShipment(ctx, sh); err != nil {
		return nil, err
	}
	for _, ev := range data.Events {
		status := ev.Status
		desc := ev.Description
		loc := ev.Location
		lat, lng := ev.Lat, ev.Lng
		e := &ShipmentEvent{ShipmentID: sh.ID, Status: &status, Description: &desc, Location: &loc, Lat: &lat, Lng: &lng, EventTime: ev.EventTime}
		_ = s.repo.CreateEvent(ctx, e)
	}
	return sh, nil
}

func (s *Service) PollActive(ctx context.Context) error {
	shipments, err := s.repo.ListActiveForPoll(ctx)
	if err != nil {
		return err
	}
	for i := range shipments {
		_, _ = s.syncOne(ctx, &shipments[i])
	}
	return nil
}

func (s *Service) HandleWebhook(ctx context.Context, payload WebhookPayload) error {
	return s.repo.UpsertFromWebhook(ctx, payload)
}

func (s *Service) ListMapPoints(ctx context.Context) ([]MapPoint, error) {
	return s.repo.ListMapPoints(ctx)
}
