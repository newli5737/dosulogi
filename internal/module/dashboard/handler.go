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
	OpenTickets   int     `json:"open_tickets"`
	ActiveOpps    int     `json:"active_opportunities"`
	PaidInvoices  int     `json:"paid_invoices"`
}

type FunnelStage struct {
	Stage string  `json:"stage"`
	Count int     `json:"count"`
	Value float64 `json:"value"`
}

type TrendPoint struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func (h *Handler) summary(ctx context.Context) (*Summary, error) {
	s := &Summary{}
	_ = h.db.QueryRow(ctx, `SELECT COALESCE(SUM(total),0) FROM invoices WHERE status='paid'`).Scan(&s.Revenue)
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM shipments`).Scan(&s.ShipmentCount)
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM customers WHERE created_at >= date_trunc('month', now())`).Scan(&s.NewCustomers)
	_ = h.db.QueryRow(ctx, `SELECT COALESCE(SUM(total),0) FROM invoices WHERE status IN ('sent','overdue')`).Scan(&s.TotalAR)
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM tickets WHERE status NOT IN ('closed','resolved')`).Scan(&s.OpenTickets)
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM opportunities WHERE stage NOT IN ('won','lost')`).Scan(&s.ActiveOpps)
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM invoices WHERE status='paid'`).Scan(&s.PaidInvoices)
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

func (h *Handler) revenueTrend(ctx context.Context) ([]TrendPoint, error) {
	rows, err := h.db.Query(ctx, `
		SELECT to_char(date_trunc('month', paid_at), 'MM/YYYY'), COALESCE(SUM(total),0)
		FROM invoices
		WHERE status = 'paid' AND paid_at >= date_trunc('month', now()) - interval '5 months'
		GROUP BY 1
		ORDER BY MIN(paid_at)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []TrendPoint
	for rows.Next() {
		var p TrendPoint
		_ = rows.Scan(&p.Label, &p.Amount)
		list = append(list, p)
	}
	return list, nil
}

func (h *Handler) ticketStats(ctx context.Context) ([]StatusCount, error) {
	rows, err := h.db.Query(ctx, `SELECT status, COUNT(*) FROM tickets GROUP BY status ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []StatusCount
	for rows.Next() {
		var s StatusCount
		_ = rows.Scan(&s.Status, &s.Count)
		list = append(list, s)
	}
	return list, nil
}

func (h *Handler) shipmentStats(ctx context.Context) ([]StatusCount, error) {
	rows, err := h.db.Query(ctx, `SELECT status, COUNT(*) FROM shipments GROUP BY status ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []StatusCount
	for rows.Next() {
		var s StatusCount
		_ = rows.Scan(&s.Status, &s.Count)
		list = append(list, s)
	}
	return list, nil
}
