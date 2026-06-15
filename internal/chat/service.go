package chat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/chat/adapter/fbchat"
	"github.com/dosu-logi/logistics-erp/internal/chat/adapter/zalo"
	"github.com/dosu-logi/logistics-erp/internal/chat/domain"
	"github.com/google/uuid"
)

type Service struct {
	repo  *Repository
	zalo  *zalo.Client
}

func NewService(repo *Repository, zaloClient *zalo.Client) *Service {
	return &Service{repo: repo, zalo: zaloClient}
}

func (s *Service) ListAccounts(ctx context.Context, platform string) ([]domain.ChatAccount, error) {
	rows, err := s.repo.ListAccounts(ctx, platform)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ChatAccount, 0, len(rows))
	for _, r := range rows {
		out = append(out, accountToDTO(r))
	}
	return out, nil
}

func (s *Service) CreateAccount(ctx context.Context, platform, name, cookiesJSON string, createdBy *uuid.UUID) (*domain.ChatAccount, error) {
	var cookiesPtr, bridgeID, externalID *string
	switch platform {
	case "facebook":
		if cookiesJSON == "" {
			return nil, errors.New("cookies_json required for facebook")
		}
		fbID, err := fbchat.VerifySession(cookiesJSON)
		if err != nil {
			return nil, fmt.Errorf("facebook session invalid: %w", err)
		}
		cookiesPtr = &cookiesJSON
		externalID = &fbID
	case "zalo":
		if !s.zalo.Available() {
			return nil, errors.New("zalo bridge not configured (set ZALO_BRIDGE_URL)")
		}
		id, err := s.zalo.CreateAccount(name)
		if err != nil {
			return nil, err
		}
		bridgeID = &id
	default:
		return nil, errors.New("unsupported platform")
	}
	row, err := s.repo.CreateAccount(ctx, platform, name, cookiesPtr, bridgeID, externalID, createdBy)
	if err != nil {
		return nil, err
	}
	dto := accountToDTO(*row)
	return &dto, nil
}

func (s *Service) UpdateAccount(ctx context.Context, id uuid.UUID, name, cookiesJSON *string) (*domain.ChatAccount, error) {
	acc, err := s.repo.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	var externalID *string
	if acc.Platform == "facebook" && cookiesJSON != nil && *cookiesJSON != "" {
		fbID, err := fbchat.VerifySession(*cookiesJSON)
		if err != nil {
			return nil, fmt.Errorf("facebook session invalid: %w", err)
		}
		externalID = &fbID
	}
	if err := s.repo.UpdateAccount(ctx, id, name, nil, cookiesJSON, externalID); err != nil {
		return nil, err
	}
	row, err := s.repo.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := accountToDTO(*row)
	return &dto, nil
}

func (s *Service) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteAccount(ctx, id)
}

func (s *Service) ZaloQRLogin(ctx context.Context, accountID uuid.UUID) (map[string]interface{}, error) {
	acc, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if acc.Platform != "zalo" || acc.ZaloBridgeID == nil {
		return nil, errors.New("not a zalo account")
	}
	return s.zalo.StartQRLogin(*acc.ZaloBridgeID)
}

func (s *Service) ZaloLoginStatus(ctx context.Context, accountID uuid.UUID) (map[string]interface{}, error) {
	acc, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if acc.Platform != "zalo" || acc.ZaloBridgeID == nil {
		return nil, errors.New("not a zalo account")
	}
	return s.zalo.LoginStatus(*acc.ZaloBridgeID)
}

func (s *Service) FetchInbox(ctx context.Context, accountID uuid.UUID, userRole, userID string) (*domain.ChatInboxResponse, error) {
	acc, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	var inbox *domain.ChatInboxResponse
	switch acc.Platform {
	case "facebook":
		if acc.CookiesJSON == nil || *acc.CookiesJSON == "" {
			return nil, errors.New("account has no cookies")
		}
		inbox, err = fbchat.FetchInbox(*acc.CookiesJSON)
	case "zalo":
		if acc.ZaloBridgeID == nil {
			return nil, errors.New("zalo account not linked")
		}
		inbox, err = s.zalo.FetchInbox(*acc.ZaloBridgeID)
	default:
		return nil, errors.New("unsupported platform")
	}
	if err != nil {
		return nil, err
	}
	_ = s.repo.TouchAccountSync(ctx, accountID)
	meta, _ := s.repo.ListConversationsByAccount(ctx, accountID)
	s.applyInboxMeta(inbox, meta, acc.Platform, accountID.String(), userRole, userID)
	return inbox, nil
}

func (s *Service) FetchThread(ctx context.Context, accountID uuid.UUID, threadID, cursor, userRole, userID string) (*domain.ChatThreadResponse, error) {
	acc, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	conv, err := s.repo.GetOrCreateConversation(ctx, acc.Platform, accountID, threadID)
	if err != nil {
		return nil, err
	}
	if !canAccessConversation(userRole, userID, conv) {
		return nil, errors.New("conversation assigned to another user")
	}
	var thread *domain.ChatThreadResponse
	switch acc.Platform {
	case "facebook":
		if acc.CookiesJSON == nil {
			return nil, errors.New("account has no cookies")
		}
		thread, err = fbchat.FetchThread(*acc.CookiesJSON, threadID, cursor)
	case "zalo":
		if acc.ZaloBridgeID == nil {
			return nil, errors.New("zalo account not linked")
		}
		thread, err = s.zalo.FetchThread(*acc.ZaloBridgeID, threadID, cursor)
	default:
		return nil, errors.New("unsupported platform")
	}
	if err != nil {
		return nil, err
	}
	for i := range thread.Messages {
		thread.Messages[i].IsSelf = thread.Messages[i].SenderFbid == thread.ViewerID
	}
	return thread, nil
}

func (s *Service) SendMessage(ctx context.Context, accountID uuid.UUID, threadID, text, userRole, userID string) error {
	acc, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}
	conv, err := s.repo.GetOrCreateConversation(ctx, acc.Platform, accountID, threadID)
	if err != nil {
		return err
	}
	if !canAccessConversation(userRole, userID, conv) {
		return errors.New("conversation assigned to another user")
	}
	switch acc.Platform {
	case "facebook":
		if acc.CookiesJSON == nil {
			return errors.New("account has no cookies")
		}
		return fbchat.SendMessage(*acc.CookiesJSON, threadID, text)
	case "zalo":
		if acc.ZaloBridgeID == nil {
			return errors.New("zalo account not linked")
		}
		return s.zalo.SendMessage(*acc.ZaloBridgeID, threadID, text)
	default:
		return errors.New("unsupported platform")
	}
}

func (s *Service) GetConversation(ctx context.Context, id uuid.UUID) (*domain.ConversationMeta, error) {
	row, err := s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, err
	}
	return conversationToMeta(*row), nil
}

func (s *Service) UpdateConversation(ctx context.Context, id uuid.UUID, customerID, assignedUserID *uuid.UUID) (*domain.ConversationMeta, error) {
	if err := s.repo.UpdateConversation(ctx, id, customerID, assignedUserID, nil, nil, nil, nil, nil); err != nil {
		return nil, err
	}
	row, err := s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, err
	}
	return conversationToMeta(*row), nil
}

func (s *Service) ListAssignees(ctx context.Context) ([]map[string]string, error) {
	rows, err := s.repo.ListAssignees(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]string{"id": r.ID.String(), "full_name": r.FullName})
	}
	return out, nil
}

func (s *Service) applyInboxMeta(inbox *domain.ChatInboxResponse, meta map[string]ConversationRow, platform, accountID, userRole, userID string) {
	filtered := inbox.Threads[:0]
	for _, t := range inbox.Threads {
		key := t.ThreadID
		if key == "" {
			key = t.ThreadKey
		}
		m, ok := meta[key]
		if !ok {
			conv, _ := s.repo.GetOrCreateConversation(context.Background(), platform, uuid.MustParse(accountID), key)
			if conv != nil {
				m = *conv
				meta[key] = m
			}
		}
		if m.AssignedUserID != nil && userRole != "admin" && userRole != "director" && m.AssignedUserID.String() != userID {
			continue
		}
		if m.AssignedUserID != nil {
			id := m.AssignedUserID.String()
			t.AssignedUserID = &id
			t.AssignedName = m.AssignedName
		}
		if m.CustomerID != nil {
			id := m.CustomerID.String()
			t.CustomerID = &id
			t.CustomerName = m.CustomerName
		}
		t.ConversationID = m.ID.String()
		t.Platform = platform
		t.AccountID = accountID
		filtered = append(filtered, t)
	}
	inbox.Threads = filtered
}

func canAccessConversation(role, userID string, conv *ConversationRow) bool {
	if role == "admin" || role == "director" {
		return true
	}
	if conv.AssignedUserID == nil {
		return true
	}
	return conv.AssignedUserID.String() == userID
}

func accountToDTO(r AccountRow) domain.ChatAccount {
	d := domain.ChatAccount{
		ID:       r.ID.String(),
		Platform: r.Platform,
		Name:     r.Name,
		Status:   r.Status,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		HasCredentials: (r.CookiesJSON != nil && *r.CookiesJSON != "") || (r.ZaloBridgeID != nil && *r.ZaloBridgeID != ""),
	}
	if r.ExternalID != nil {
		d.ExternalID = *r.ExternalID
	}
	if r.LastSyncAt != nil {
		s := r.LastSyncAt.Format(time.RFC3339)
		d.LastSyncAt = &s
	}
	return d
}

func conversationToMeta(r ConversationRow) *domain.ConversationMeta {
	m := &domain.ConversationMeta{
		ID:          r.ID.String(),
		Platform:    r.Platform,
		AccountID:   r.AccountID.String(),
		ThreadID:    r.ThreadID,
		UnreadCount: r.UnreadCount,
		CustomerName: r.CustomerName,
		AssignedName: r.AssignedName,
	}
	if r.ThreadTitle != nil {
		m.ThreadTitle = *r.ThreadTitle
	}
	if r.PeerName != nil {
		m.PeerName = *r.PeerName
	}
	if r.LastMessage != nil {
		m.LastMessage = *r.LastMessage
	}
	if r.LastMessageAt != nil {
		s := r.LastMessageAt.Format(time.RFC3339)
		m.LastMessageAt = &s
	}
	if r.CustomerID != nil {
		id := r.CustomerID.String()
		m.CustomerID = &id
	}
	if r.AssignedUserID != nil {
		id := r.AssignedUserID.String()
		m.AssignedUserID = &id
	}
	return m
}
