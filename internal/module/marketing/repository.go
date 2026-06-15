package marketing

import (
	"context"
	"errors"
	"fmt"

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

func (r *Repository) Create(ctx context.Context, c *Campaign) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO campaigns (name, type, status, subject, body_html, segment, scheduled_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, sent_count, created_at`,
		c.Name, c.Type, c.Status, c.Subject, c.BodyHTML, c.Segment, c.ScheduledAt, c.CreatedBy,
	).Scan(&c.ID, &c.SentCount, &c.CreatedAt)
}

func (r *Repository) List(ctx context.Context, status string, limit, offset int) ([]Campaign, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	n := 1
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM campaigns "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT id, name, type, status, subject, body_html, segment, scheduled_at, sent_count, created_by, created_at
		FROM campaigns %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Campaign
	for rows.Next() {
		var c Campaign
		_ = rows.Scan(&c.ID, &c.Name, &c.Type, &c.Status, &c.Subject, &c.BodyHTML, &c.Segment, &c.ScheduledAt, &c.SentCount, &c.CreatedBy, &c.CreatedAt)
		list = append(list, c)
	}
	return list, total, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*Campaign, error) {
	c := &Campaign{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, type, status, subject, body_html, segment, scheduled_at, sent_count, created_by, created_at
		FROM campaigns WHERE id=$1`, id,
	).Scan(&c.ID, &c.Name, &c.Type, &c.Status, &c.Subject, &c.BodyHTML, &c.Segment, &c.ScheduledAt, &c.SentCount, &c.CreatedBy, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *Repository) Update(ctx context.Context, c *Campaign) error {
	_, err := r.db.Exec(ctx, `
		UPDATE campaigns SET name=$1, type=$2, status=$3, subject=$4, body_html=$5, segment=$6, scheduled_at=$7
		WHERE id=$8`, c.Name, c.Type, c.Status, c.Subject, c.BodyHTML, c.Segment, c.ScheduledAt, c.ID)
	return err
}

func (r *Repository) IncrementSentCount(ctx context.Context, id uuid.UUID, count int) error {
	_, err := r.db.Exec(ctx, `UPDATE campaigns SET sent_count = sent_count + $1, status='done' WHERE id=$2`, count, id)
	return err
}

func (r *Repository) ListTargetEmails(ctx context.Context, segmentJSON []byte) ([]struct {
	ID    uuid.UUID
	Email string
}, error) {
	// Simple segment filter: if empty, send to all customers with email
	q := `SELECT id, email FROM customers WHERE email IS NOT NULL AND email != ''`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []struct {
		ID    uuid.UUID
		Email string
	}
	for rows.Next() {
		var item struct {
			ID    uuid.UUID
			Email string
		}
		_ = rows.Scan(&item.ID, &item.Email)
		list = append(list, item)
	}
	_ = segmentJSON
	return list, nil
}

func (r *Repository) CreateLog(ctx context.Context, log *CampaignLog) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO campaign_logs (campaign_id, customer_id, email, status, sg_message_id)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		log.CampaignID, log.CustomerID, log.Email, log.Status, log.SGMessageID,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *Repository) ListLogs(ctx context.Context, campaignID uuid.UUID, status string, limit, offset int) ([]CampaignLog, int, error) {
	where := "WHERE campaign_id = $1"
	args := []interface{}{campaignID}
	n := 2
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM campaign_logs "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT id, campaign_id, customer_id, email, status, sg_message_id, created_at
		FROM campaign_logs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, n, n+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []CampaignLog
	for rows.Next() {
		var l CampaignLog
		_ = rows.Scan(&l.ID, &l.CampaignID, &l.CustomerID, &l.Email, &l.Status, &l.SGMessageID, &l.CreatedAt)
		list = append(list, l)
	}
	return list, total, nil
}

func (r *Repository) UpdateLogByMessageID(ctx context.Context, sgMessageID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE campaign_logs SET status=$1 WHERE sg_message_id=$2`, status, sgMessageID)
	return err
}

func (r *Repository) ListScheduledDue(ctx context.Context) ([]Campaign, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, type, status, subject, body_html, segment, scheduled_at, sent_count, created_by, created_at
		FROM campaigns WHERE status='scheduled' AND scheduled_at <= now()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Campaign
	for rows.Next() {
		var c Campaign
		_ = rows.Scan(&c.ID, &c.Name, &c.Type, &c.Status, &c.Subject, &c.BodyHTML, &c.Segment, &c.ScheduledAt, &c.SentCount, &c.CreatedBy, &c.CreatedAt)
		list = append(list, c)
	}
	return list, nil
}
