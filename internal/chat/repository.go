package chat

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRow struct {
	ID             uuid.UUID
	Platform       string
	Name           string
	ExternalID     *string
	CookiesJSON    *string
	ZaloBridgeID   *string
	Status         string
	LastSyncAt     *time.Time
	CreatedBy      *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ConversationRow struct {
	ID             uuid.UUID
	Platform       string
	AccountID      uuid.UUID
	ThreadID       string
	ThreadTitle    *string
	PeerName       *string
	PeerAvatar     *string
	CustomerID     *uuid.UUID
	AssignedUserID *uuid.UUID
	LastMessage    *string
	LastMessageAt  *time.Time
	UnreadCount    int
	CustomerName   string
	AssignedName   string
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListAccounts(ctx context.Context, platform string) ([]AccountRow, error) {
	q := `SELECT id, platform, name, external_id, cookies_json, zalo_bridge_id, status, last_sync_at, created_by, created_at, updated_at
		FROM chat_accounts WHERE 1=1`
	args := []interface{}{}
	if platform != "" {
		q += ` AND platform = $1`
		args = append(args, platform)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AccountRow
	for rows.Next() {
		var a AccountRow
		if err := rows.Scan(&a.ID, &a.Platform, &a.Name, &a.ExternalID, &a.CookiesJSON, &a.ZaloBridgeID, &a.Status, &a.LastSyncAt, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (r *Repository) GetAccount(ctx context.Context, id uuid.UUID) (*AccountRow, error) {
	var a AccountRow
	err := r.db.QueryRow(ctx, `
		SELECT id, platform, name, external_id, cookies_json, zalo_bridge_id, status, last_sync_at, created_by, created_at, updated_at
		FROM chat_accounts WHERE id = $1`, id).Scan(
		&a.ID, &a.Platform, &a.Name, &a.ExternalID, &a.CookiesJSON, &a.ZaloBridgeID, &a.Status, &a.LastSyncAt, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) CreateAccount(ctx context.Context, platform, name string, cookiesJSON, zaloBridgeID, externalID *string, createdBy *uuid.UUID) (*AccountRow, error) {
	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		INSERT INTO chat_accounts (id, platform, name, external_id, cookies_json, zalo_bridge_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, platform, name, externalID, cookiesJSON, zaloBridgeID, createdBy,
	)
	if err != nil {
		return nil, err
	}
	return r.GetAccount(ctx, id)
}

func (r *Repository) UpdateAccount(ctx context.Context, id uuid.UUID, name, status, cookiesJSON, externalID *string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE chat_accounts SET
			name = COALESCE($2, name),
			status = COALESCE($3, status),
			cookies_json = COALESCE($4, cookies_json),
			external_id = COALESCE($5, external_id),
			updated_at = now()
		WHERE id = $1`, id, name, status, cookiesJSON, externalID)
	return err
}

func (r *Repository) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM chat_accounts WHERE id = $1`, id)
	return err
}

func (r *Repository) TouchAccountSync(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE chat_accounts SET last_sync_at = now(), updated_at = now() WHERE id = $1`, id)
	return err
}

func (r *Repository) GetOrCreateConversation(ctx context.Context, platform string, accountID uuid.UUID, threadID string) (*ConversationRow, error) {
	var c ConversationRow
	err := r.db.QueryRow(ctx, `
		SELECT c.id, c.platform, c.account_id, c.thread_id, c.thread_title, c.peer_name, c.peer_avatar,
			c.customer_id, c.assigned_user_id, c.last_message, c.last_message_at, c.unread_count,
			COALESCE(cu.name, ''), COALESCE(u.full_name, '')
		FROM chat_conversations c
		LEFT JOIN customers cu ON cu.id = c.customer_id
		LEFT JOIN users u ON u.id = c.assigned_user_id
		WHERE c.platform = $1 AND c.account_id = $2 AND c.thread_id = $3`,
		platform, accountID, threadID,
	).Scan(&c.ID, &c.Platform, &c.AccountID, &c.ThreadID, &c.ThreadTitle, &c.PeerName, &c.PeerAvatar,
		&c.CustomerID, &c.AssignedUserID, &c.LastMessage, &c.LastMessageAt, &c.UnreadCount, &c.CustomerName, &c.AssignedName)
	if err == nil {
		return &c, nil
	}
	id := uuid.New()
	_, err = r.db.Exec(ctx, `
		INSERT INTO chat_conversations (id, platform, account_id, thread_id)
		VALUES ($1, $2, $3, $4)`, id, platform, accountID, threadID)
	if err != nil {
		return nil, err
	}
	c = ConversationRow{ID: id, Platform: platform, AccountID: accountID, ThreadID: threadID}
	return &c, nil
}

func (r *Repository) ListConversationsByAccount(ctx context.Context, accountID uuid.UUID) (map[string]ConversationRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.platform, c.account_id, c.thread_id, c.thread_title, c.peer_name, c.peer_avatar,
			c.customer_id, c.assigned_user_id, c.last_message, c.last_message_at, c.unread_count,
			COALESCE(cu.name, ''), COALESCE(u.full_name, '')
		FROM chat_conversations c
		LEFT JOIN customers cu ON cu.id = c.customer_id
		LEFT JOIN users u ON u.id = c.assigned_user_id
		WHERE c.account_id = $1`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]ConversationRow{}
	for rows.Next() {
		var c ConversationRow
		if err := rows.Scan(&c.ID, &c.Platform, &c.AccountID, &c.ThreadID, &c.ThreadTitle, &c.PeerName, &c.PeerAvatar,
			&c.CustomerID, &c.AssignedUserID, &c.LastMessage, &c.LastMessageAt, &c.UnreadCount, &c.CustomerName, &c.AssignedName); err != nil {
			return nil, err
		}
		out[c.ThreadID] = c
	}
	return out, nil
}

func (r *Repository) UpdateConversation(ctx context.Context, id uuid.UUID, customerID, assignedUserID *uuid.UUID, threadTitle, peerName, peerAvatar, lastMessage *string, lastMessageAt *time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE chat_conversations SET
			customer_id = COALESCE($2, customer_id),
			assigned_user_id = COALESCE($3, assigned_user_id),
			thread_title = COALESCE($4, thread_title),
			peer_name = COALESCE($5, peer_name),
			peer_avatar = COALESCE($6, peer_avatar),
			last_message = COALESCE($7, last_message),
			last_message_at = COALESCE($8, last_message_at),
			updated_at = now()
		WHERE id = $1`, id, customerID, assignedUserID, threadTitle, peerName, peerAvatar, lastMessage, lastMessageAt)
	return err
}

func (r *Repository) GetConversation(ctx context.Context, id uuid.UUID) (*ConversationRow, error) {
	var c ConversationRow
	err := r.db.QueryRow(ctx, `
		SELECT c.id, c.platform, c.account_id, c.thread_id, c.thread_title, c.peer_name, c.peer_avatar,
			c.customer_id, c.assigned_user_id, c.last_message, c.last_message_at, c.unread_count,
			COALESCE(cu.name, ''), COALESCE(u.full_name, '')
		FROM chat_conversations c
		LEFT JOIN customers cu ON cu.id = c.customer_id
		LEFT JOIN users u ON u.id = c.assigned_user_id
		WHERE c.id = $1`, id).Scan(
		&c.ID, &c.Platform, &c.AccountID, &c.ThreadID, &c.ThreadTitle, &c.PeerName, &c.PeerAvatar,
		&c.CustomerID, &c.AssignedUserID, &c.LastMessage, &c.LastMessageAt, &c.UnreadCount, &c.CustomerName, &c.AssignedName,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) ListAssignees(ctx context.Context) ([]struct {
	ID       uuid.UUID
	FullName string
}, error) {
	rows, err := r.db.Query(ctx, `SELECT id, full_name FROM users WHERE role IN ('admin','sales','director') AND is_active = true ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []struct {
		ID       uuid.UUID
		FullName string
	}
	for rows.Next() {
		var u struct {
			ID       uuid.UUID
			FullName string
		}
		if err := rows.Scan(&u.ID, &u.FullName); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, nil
}
