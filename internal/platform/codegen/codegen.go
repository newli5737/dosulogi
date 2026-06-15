package codegen

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Next(ctx context.Context, db *pgxpool.Pool, table, prefix string, withYear bool) (string, error) {
	if withYear {
		year := time.Now().Year()
		pattern := fmt.Sprintf("%s-%d-%%", prefix, year)
		var count int
		q := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE code LIKE $1`, table)
		if err := db.QueryRow(ctx, q, pattern).Scan(&count); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s-%d-%05d", prefix, year, count+1), nil
	}
	var count int
	q := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, table)
	if err := db.QueryRow(ctx, q).Scan(&count); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%05d", prefix, count+1), nil
}
