package tracking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, status, customerID string, limit, offset int) ([]Shipment, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	if customerID != "" {
		where += fmt.Sprintf(" AND customer_id = $%d", n)
		args = append(args, customerID)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM shipments "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT id, tracking_code, external_id, customer_id, contract_id, status, origin, destination,
		lat, lng, estimated_delivery, actual_delivery, last_synced_at, created_at, updated_at
		FROM shipments %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Shipment
	for rows.Next() {
		var s Shipment
		_ = rows.Scan(&s.ID, &s.TrackingCode, &s.ExternalID, &s.CustomerID, &s.ContractID, &s.Status,
			&s.Origin, &s.Destination, &s.Lat, &s.Lng, &s.EstimatedDelivery, &s.ActualDelivery,
			&s.LastSyncedAt, &s.CreatedAt, &s.UpdatedAt)
		list = append(list, s)
	}
	return list, total, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*Shipment, error) {
	s := &Shipment{}
	err := r.db.QueryRow(ctx, `
		SELECT id, tracking_code, external_id, customer_id, contract_id, status, origin, destination,
		lat, lng, estimated_delivery, actual_delivery, raw_payload, last_synced_at, created_at, updated_at
		FROM shipments WHERE id=$1`, id,
	).Scan(&s.ID, &s.TrackingCode, &s.ExternalID, &s.CustomerID, &s.ContractID, &s.Status,
		&s.Origin, &s.Destination, &s.Lat, &s.Lng, &s.EstimatedDelivery, &s.ActualDelivery,
		&s.RawPayload, &s.LastSyncedAt, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return s, err
}

func (r *Repository) GetByTrackingCode(ctx context.Context, code string) (*Shipment, error) {
	s := &Shipment{}
	err := r.db.QueryRow(ctx, `
		SELECT id, tracking_code, external_id, customer_id, contract_id, status, origin, destination,
		lat, lng, estimated_delivery, actual_delivery, raw_payload, last_synced_at, created_at, updated_at
		FROM shipments WHERE tracking_code=$1`, code,
	).Scan(&s.ID, &s.TrackingCode, &s.ExternalID, &s.CustomerID, &s.ContractID, &s.Status,
		&s.Origin, &s.Destination, &s.Lat, &s.Lng, &s.EstimatedDelivery, &s.ActualDelivery,
		&s.RawPayload, &s.LastSyncedAt, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return s, err
}

func (r *Repository) ListActiveForPoll(ctx context.Context) ([]Shipment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tracking_code, external_id, customer_id, contract_id, status, origin, destination,
		lat, lng, estimated_delivery, actual_delivery, last_synced_at, created_at, updated_at
		FROM shipments WHERE status IS NULL OR status NOT IN ('delivered', 'failed')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Shipment
	for rows.Next() {
		var s Shipment
		_ = rows.Scan(&s.ID, &s.TrackingCode, &s.ExternalID, &s.CustomerID, &s.ContractID, &s.Status,
			&s.Origin, &s.Destination, &s.Lat, &s.Lng, &s.EstimatedDelivery, &s.ActualDelivery,
			&s.LastSyncedAt, &s.CreatedAt, &s.UpdatedAt)
		list = append(list, s)
	}
	return list, nil
}

func (r *Repository) UpdateShipment(ctx context.Context, s *Shipment) error {
	_, err := r.db.Exec(ctx, `
		UPDATE shipments SET status=$1, origin=$2, destination=$3, lat=$4, lng=$5,
		estimated_delivery=$6, actual_delivery=$7, raw_payload=$8, last_synced_at=now(), updated_at=now()
		WHERE id=$9`, s.Status, s.Origin, s.Destination, s.Lat, s.Lng, s.EstimatedDelivery, s.ActualDelivery, s.RawPayload, s.ID)
	return err
}

func (r *Repository) CreateEvent(ctx context.Context, e *ShipmentEvent) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO shipment_events (shipment_id, status, description, location, lat, lng, event_time)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
		e.ShipmentID, e.Status, e.Description, e.Location, e.Lat, e.Lng, e.EventTime,
	).Scan(&e.ID, &e.CreatedAt)
}

func (r *Repository) ListEvents(ctx context.Context, shipmentID uuid.UUID) ([]ShipmentEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, shipment_id, status, description, location, lat, lng, event_time, created_at
		FROM shipment_events WHERE shipment_id=$1 ORDER BY event_time DESC NULLS LAST`, shipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ShipmentEvent
	for rows.Next() {
		var e ShipmentEvent
		_ = rows.Scan(&e.ID, &e.ShipmentID, &e.Status, &e.Description, &e.Location, &e.Lat, &e.Lng, &e.EventTime, &e.CreatedAt)
		list = append(list, e)
	}
	return list, nil
}

func (r *Repository) ListMapPoints(ctx context.Context) ([]MapPoint, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.tracking_code, COALESCE(s.status,''), s.lat, s.lng, COALESCE(c.name,''), COALESCE(s.destination,'')
		FROM shipments s LEFT JOIN customers c ON c.id = s.customer_id
		WHERE s.lat IS NOT NULL AND s.lng IS NOT NULL
		AND (s.status IS NULL OR s.status NOT IN ('delivered', 'failed'))`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []MapPoint
	for rows.Next() {
		var p MapPoint
		_ = rows.Scan(&p.TrackingCode, &p.Status, &p.Lat, &p.Lng, &p.CustomerName, &p.Destination)
		list = append(list, p)
	}
	return list, nil
}

func (r *Repository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]Shipment, error) {
	list, _, err := r.List(ctx, "", customerID.String(), 100, 0)
	return list, err
}

func (r *Repository) UpsertFromWebhook(ctx context.Context, payload WebhookPayload) error {
	s, err := r.GetByTrackingCode(ctx, payload.TrackingCode)
	now := time.Now()
	if errors.Is(err, ErrNotFound) {
		status := payload.Status
		loc := payload.Location
		lat, lng := payload.Lat, payload.Lng
		s = &Shipment{TrackingCode: payload.TrackingCode, Status: &status, Destination: &loc, Lat: &lat, Lng: &lng, LastSyncedAt: &now}
		return r.db.QueryRow(ctx, `
			INSERT INTO shipments (tracking_code, status, destination, lat, lng, last_synced_at)
			VALUES ($1,$2,$3,$4,$5,now()) RETURNING id`, s.TrackingCode, s.Status, s.Destination, s.Lat, s.Lng,
		).Scan(&s.ID)
	}
	if err != nil {
		return err
	}
	status := payload.Status
	loc := payload.Location
	lat, lng := payload.Lat, payload.Lng
	s.Status = &status
	s.Destination = &loc
	s.Lat = &lat
	s.Lng = &lng
	if err := r.UpdateShipment(ctx, s); err != nil {
		return err
	}
	desc := payload.Description
	e := &ShipmentEvent{ShipmentID: s.ID, Status: &status, Description: &desc, Location: &loc, Lat: &lat, Lng: &lng, EventTime: payload.EventTime}
	return r.CreateEvent(ctx, e)
}
