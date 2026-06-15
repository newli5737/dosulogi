package domain

type ChatThread struct {
	ThreadID       string     `json:"thread_id"`
	ThreadKey      string     `json:"thread_key"`
	Title          string     `json:"title"`
	IsGroup        bool       `json:"is_group"`
	LastActivityMs string     `json:"last_activity_ms"`
	Users          []ChatUser `json:"users"`
	LastMessage    string     `json:"last_message"`
	AssignedUserID *string    `json:"assigned_user_id,omitempty"`
	AssignedName   string     `json:"assigned_name,omitempty"`
	CustomerID     *string    `json:"customer_id,omitempty"`
	CustomerName   string     `json:"customer_name,omitempty"`
	Platform       string     `json:"platform,omitempty"`
	AccountID      string     `json:"account_id,omitempty"`
	ConversationID string     `json:"conversation_id,omitempty"`
}

type ChatMessage struct {
	MessageID    string `json:"message_id"`
	SenderFbid   string `json:"sender_fbid"`
	SenderName   string `json:"sender_name"`
	SenderAvatar string `json:"sender_avatar"`
	Text         string `json:"text"`
	ContentType  string `json:"content_type"`
	TimestampMs  string `json:"timestamp_ms"`
	IsSelf       bool   `json:"is_self,omitempty"`
}

type ChatUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Avatar   string `json:"avatar"`
}

type ChatInboxResponse struct {
	Threads    []ChatThread `json:"threads"`
	HasMore    bool         `json:"has_more"`
	NextCursor string       `json:"next_cursor"`
	ViewerID   string       `json:"viewer_id,omitempty"`
}

type ChatThreadResponse struct {
	ThreadID   string        `json:"thread_id"`
	Title      string        `json:"title"`
	Users      []ChatUser    `json:"users"`
	Messages   []ChatMessage `json:"messages"`
	HasMore    bool          `json:"has_more"`
	NextCursor string        `json:"next_cursor"`
	ViewerID   string        `json:"viewer_id,omitempty"`
}

type ChatAccount struct {
	ID            string  `json:"id"`
	Platform      string  `json:"platform"`
	Name          string  `json:"name"`
	ExternalID    string  `json:"external_id,omitempty"`
	Status        string  `json:"status"`
	LastSyncAt    *string `json:"last_sync_at,omitempty"`
	HasCredentials bool   `json:"has_credentials"`
	CreatedAt     string  `json:"created_at"`
}

type ConversationMeta struct {
	ID             string  `json:"id"`
	Platform       string  `json:"platform"`
	AccountID      string  `json:"account_id"`
	ThreadID       string  `json:"thread_id"`
	ThreadTitle    string  `json:"thread_title,omitempty"`
	PeerName       string  `json:"peer_name,omitempty"`
	CustomerID     *string `json:"customer_id,omitempty"`
	CustomerName   string  `json:"customer_name,omitempty"`
	AssignedUserID *string `json:"assigned_user_id,omitempty"`
	AssignedName   string  `json:"assigned_name,omitempty"`
	LastMessage    string  `json:"last_message,omitempty"`
	LastMessageAt  *string `json:"last_message_at,omitempty"`
	UnreadCount    int     `json:"unread_count"`
}
