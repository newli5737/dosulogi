package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound       = errors.New("user not found")
	ErrInvalidCred    = errors.New("invalid credentials")
	ErrInactive       = errors.New("account inactive")
	ErrDuplicateEmail = errors.New("email already exists")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password, full_name, role, is_active, created_at, updated_at
		FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password, full_name, role, is_active, created_at, updated_at
		FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *Repository) List(ctx context.Context, role string, limit, offset int) ([]User, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argN := 1
	if role != "" {
		where += fmt.Sprintf(" AND role = $%d", argN)
		args = append(args, role)
		argN++
	}

	var total int
	countQ := "SELECT COUNT(*) FROM users " + where
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	q := fmt.Sprintf(`SELECT id, email, password, full_name, role, is_active, created_at, updated_at
		FROM users %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argN, argN+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *Repository) Create(ctx context.Context, u *User) error {
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email, password, full_name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`,
		u.Email, u.Password, u.FullName, u.Role,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil && isUniqueViolation(err) {
		return ErrDuplicateEmail
	}
	return err
}

func (r *Repository) Update(ctx context.Context, u *User) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET full_name=$1, role=$2, is_active=$3, updated_at=now()
		WHERE id=$4`, u.FullName, u.Role, u.IsActive, u.ID)
	return err
}

func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET password=$1, updated_at=now() WHERE id=$2`, hash, id)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	return err
}

func (r *Repository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`, userID, tokenHash, expiresAt)
	return err
}

func (r *Repository) FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	rt := &RefreshToken{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked
		FROM refresh_tokens WHERE token_hash=$1 AND revoked=false AND expires_at > now()`,
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.Revoked)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("invalid refresh token")
	}
	return rt, err
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_tokens SET revoked=true WHERE token_hash=$1`, tokenHash)
	return err
}

func (r *Repository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_tokens SET revoked=true WHERE user_id=$1`, userID)
	return err
}

func isUniqueViolation(err error) bool {
	return err != nil && (err.Error() == "duplicate key value violates unique constraint" ||
		contains(err.Error(), "unique"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && searchSub(s, sub))
}

func searchSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
