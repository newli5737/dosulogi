package dashboard

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

type Summary struct {
	Revenue       float64 `json:"revenue"`
	ShipmentCount int     `json:"shipment_count"`
	NewCustomers  int     `json:"new_customers"`
	TotalAR       float64 `json:"total_ar"`
}

type FunnelStage struct {
	Stage string  `json:"stage"`
	Count int     `json:"count"`
	Value float64 `json:"value"`
}

func (h *Handler) summary(ctx context.Context) (*Summary, error) {
	s := &Summary{}
	_ = h.db.QueryRow(ctx, `SELECT COALESCE(SUM(total),0) FROM invoices WHERE status='paid'`).Scan(&s.Revenue)
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM shipments`).Scan(&s.ShipmentCount)
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM customers WHERE created_at >= date_trunc('month', now())`).Scan(&s.NewCustomers)
	_ = h.db.QueryRow(ctx, `SELECT COALESCE(SUM(total),0) FROM invoices WHERE status IN ('sent','overdue')`).Scan(&s.TotalAR)
	return s, nil
}

func (h *Handler) salesFunnel(ctx context.Context) ([]FunnelStage, error) {
	rows, err := h.db.Query(ctx, `
		SELECT stage, COUNT(*), COALESCE(SUM(value),0)
		FROM opportunities GROUP BY stage ORDER BY
		CASE stage
			WHEN 'lead' THEN 1 WHEN 'qualified' THEN 2 WHEN 'proposal' THEN 3
			WHEN 'negotiation' THEN 4 WHEN 'won' THEN 5 WHEN 'lost' THEN 6 ELSE 7 END`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []FunnelStage
	for rows.Next() {
		var f FunnelStage
		_ = rows.Scan(&f.Stage, &f.Count, &f.Value)
		list = append(list, f)
	}
	return list, nil
}
