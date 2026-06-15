package fbchat

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/chat/domain"
)

// ── Session Management ──────────────────────────────────────────────────────

const (
	gqlSessionTTL = 30 * time.Minute
)

type gqlSession struct {
	tokens     *fbTokens
	cookieStr  string
	lastUsedAt time.Time
}

var (
	gqlSessionsMu sync.Mutex
	gqlSessions   = map[string]*gqlSession{}
)

func gqlSessionKey(cookieStr string) string {
	sum := sha1.Sum([]byte(cookieStr))
	return hex.EncodeToString(sum[:])
}

func getGQLSession(cookieStr string) (*gqlSession, error) {
	key := gqlSessionKey(cookieStr)
	now := time.Now()

	gqlSessionsMu.Lock()
	sess := gqlSessions[key]
	if sess != nil && now.Sub(sess.lastUsedAt) > gqlSessionTTL {
		delete(gqlSessions, key)
		sess = nil
	}
	gqlSessionsMu.Unlock()

	if sess == nil {
		tokens, err := fetchFBTokens(cookieStr)
		if err != nil {
			return nil, fmt.Errorf("gql session init: %w", err)
		}
		sess = &gqlSession{
			tokens:     tokens,
			cookieStr:  cookieStr,
			lastUsedAt: now,
		}
		gqlSessionsMu.Lock()
		gqlSessions[key] = sess
		gqlSessionsMu.Unlock()
	}

	gqlSessionsMu.Lock()
	sess.lastUsedAt = now
	gqlSessionsMu.Unlock()

	return sess, nil
}

// ── GraphQL Helpers ─────────────────────────────────────────────────────────

// graphqlBatchPost posts a GraphQL batch request (same as Python _state._post with as_graphql=True)
func graphqlBatchPost(sess *gqlSession, queries map[string]interface{}) ([]byte, error) {
	queriesJSON, _ := json.Marshal(queries)

	form := url.Values{
		"fb_dtsg": {sess.tokens.FbDtsg},
		"jazoest": {sess.tokens.Jazoest},
		"__a":     {"1"},
		"__user":  {sess.tokens.FacebookID},
		"av":      {sess.tokens.FacebookID},
		"__rev":   {sess.tokens.ClientRevision},
		"queries": {string(queriesJSON)},
	}

	req, _ := http.NewRequest("POST", fbBaseURL+"/api/graphqlbatch/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", sess.cookieStr)
	req.Header.Set("User-Agent", fbUA)
	req.Header.Set("Origin", fbBaseURL)
	req.Header.Set("Referer", fbBaseURL+"/messages/")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graphql batch request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphql batch status %d: %s", resp.StatusCode, string(body[:min(len(body), 300)]))
	}

	return body, nil
}

// graphqlPost posts a single GraphQL request (same as Python _state._post to /api/graphql/)
func graphqlPost(sess *gqlSession, data url.Values) ([]byte, error) {
	// Add standard params
	data.Set("__a", "1")
	data.Set("__user", sess.tokens.FacebookID)
	data.Set("fb_dtsg", sess.tokens.FbDtsg)
	data.Set("jazoest", sess.tokens.Jazoest)

	req, _ := http.NewRequest("POST", fbBaseURL+"/api/graphql/", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", sess.cookieStr)
	req.Header.Set("User-Agent", fbUA)
	req.Header.Set("Origin", fbBaseURL)
	req.Header.Set("Referer", fbBaseURL+"/messages/")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("X-Fb-Lsd", sess.tokens.FbDtsg)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graphql request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphql status %d", resp.StatusCode)
	}

	return body, nil
}

// stripAndParseBatch strips "for (;;);" and parses concatenated JSON objects
func stripAndParseBatch(body []byte) (map[string]interface{}, error) {
	text := string(body)
	if idx := strings.Index(text, "for (;;);"); idx >= 0 {
		text = text[idx+9:]
	}
	return parseGraphQLBatch(text)
}

// ── Fetch Inbox (port of Python fetch_thread_list) ──────────────────────────

// FetchInboxGQL fetches inbox using GraphQL batch API.
// Port of Python: MessengerClient.fetch_thread_list(limit=50, thread_folder=INBOX)
// Uses doc_id=3426149104143726
func FetchInboxGQL(cookieStr string) (*domain.ChatInboxResponse, error) {
	sess, err := getGQLSession(cookieStr)
	if err != nil {
		return nil, err
	}

	queries := map[string]interface{}{
		"o0": map[string]interface{}{
			"doc_id": "3426149104143726",
			"query_params": map[string]interface{}{
				"limit":                  50,
				"tags":                   []string{"INBOX"},
				"before":                 nil,
				"includeDeliveryReceipts": true,
				"includeSeqID":           false,
			},
		},
	}

	body, err := graphqlBatchPost(sess, queries)
	if err != nil {
		return nil, err
	}

	parsed, err := stripAndParseBatch(body)
	if err != nil {
		// Fallback: try old doc_id
		queries["o0"].(map[string]interface{})["doc_id"] = "3336396659757871"
		body2, err2 := graphqlBatchPost(sess, queries)
		if err2 != nil {
			return nil, fmt.Errorf("both doc_ids failed: %w / %w", err, err2)
		}
		parsed, err = stripAndParseBatch(body2)
		if err != nil {
			return nil, err
		}
	}

	return parseFBInbox(parsed, sess.tokens.FacebookID)
}

// ── Fetch Thread Messages (port of Python fetch_thread_messages) ─────────────

// FetchThreadGQL fetches messages for a thread using GraphQL batch API.
// Port of Python: MessengerClient.fetch_thread_messages(thread_id, message_limit=20)
// Uses doc_id=1860982147341344
func FetchThreadGQL(cookieStr string, threadID string, cursor string) (*domain.ChatThreadResponse, error) {
	sess, err := getGQLSession(cookieStr)
	if err != nil {
		return nil, err
	}

	messageLimit := 20
	var beforeTS interface{} = nil

	if cursor != "" {
		cur, err := decodeGQLCursor(cursor)
		if err == nil && cur.Before > 0 {
			beforeTS = cur.Before
		}
	}

	// Port of Python fetch_thread_messages: doc_id=1860982147341344
	queries := map[string]interface{}{
		"o0": map[string]interface{}{
			"doc_id": "1860982147341344",
			"query_params": map[string]interface{}{
				"id":                 threadID,
				"message_limit":      messageLimit,
				"load_messages":      true,
				"load_read_receipts": true,
				"before":             beforeTS,
			},
		},
	}

	body, err := graphqlBatchPost(sess, queries)
	if err != nil {
		return nil, err
	}

	parsed, err := stripAndParseBatch(body)
	if err != nil {
		return nil, err
	}

	result, err := parseThreadMessages(parsed, sess.tokens.FacebookID, threadID)
	if err != nil {
		return nil, err
	}

	slog.Info("gql thread fetched",
		"thread_id", threadID,
		"messages", len(result.Messages),
		"users", len(result.Users),
		"has_more", result.HasMore,
	)

	return result, nil
}

// parseThreadMessages parses the GraphQL response for thread messages.
// Response structure: o0 -> data -> message_thread -> messages -> nodes[]
func parseThreadMessages(data map[string]interface{}, viewerID, threadID string) (*domain.ChatThreadResponse, error) {
	o0, _ := data["o0"].(map[string]interface{})
	if o0 == nil {
		return nil, fmt.Errorf("missing o0 in thread response")
	}
	d, _ := o0["data"].(map[string]interface{})
	if d == nil {
		return nil, fmt.Errorf("missing data in o0")
	}
	msgThread, _ := d["message_thread"].(map[string]interface{})
	if msgThread == nil {
		return nil, fmt.Errorf("missing message_thread")
	}

	result := &domain.ChatThreadResponse{
		ThreadID: threadID,
		ViewerID: viewerID,
	}

	// Thread name
	result.Title = getString(msgThread, "name")

	// Thread type
	threadType := getString(msgThread, "thread_type")
	_ = threadType

	// Parse participants
	allP, _ := msgThread["all_participants"].(map[string]interface{})
	if allP != nil {
		edges, _ := allP["edges"].([]interface{})
		for _, e := range edges {
			edge, _ := e.(map[string]interface{})
			if edge == nil {
				continue
			}
			pNode, _ := edge["node"].(map[string]interface{})
			if pNode == nil {
				continue
			}
			actor, _ := pNode["messaging_actor"].(map[string]interface{})
			if actor == nil {
				continue
			}
			user := domain.ChatUser{
				ID:       getString(actor, "id"),
				Username: getString(actor, "username"),
				FullName: getString(actor, "name"),
			}
			if bi, _ := actor["big_image_src"].(map[string]interface{}); bi != nil {
				user.Avatar = getString(bi, "uri")
			}
			result.Users = append(result.Users, user)
		}
	}

	// Set title from other participant for 1:1 chats
	if result.Title == "" && len(result.Users) >= 2 {
		for _, u := range result.Users {
			if u.ID != viewerID {
				result.Title = u.FullName
				if result.Title == "" {
					result.Title = u.Username
				}
				break
			}
		}
	}

	// Build sender map for quick lookup
	senderMap := map[string]domain.ChatUser{}
	for _, u := range result.Users {
		senderMap[u.ID] = u
	}

	// Parse messages
	messages, _ := msgThread["messages"].(map[string]interface{})
	if messages != nil {
		msgNodes, _ := messages["nodes"].([]interface{})
		for _, mn := range msgNodes {
			msgNode, ok := mn.(map[string]interface{})
			if !ok {
				continue
			}

			msg := domain.ChatMessage{
				MessageID:   getString(msgNode, "message_id"),
				TimestampMs: getString(msgNode, "timestamp_precise"),
			}

			// Text from message.text or snippet
			if msgObj, _ := msgNode["message"].(map[string]interface{}); msgObj != nil {
				msg.Text = getString(msgObj, "text")
			}
			if msg.Text == "" {
				msg.Text = getString(msgNode, "snippet")
			}

			// Sender
			if sender, _ := msgNode["message_sender"].(map[string]interface{}); sender != nil {
				if actor, _ := sender["messaging_actor"].(map[string]interface{}); actor != nil {
					msg.SenderFbid = getString(actor, "id")
					msg.SenderName = getString(actor, "name")
					if bi, _ := actor["big_image_src"].(map[string]interface{}); bi != nil {
						msg.SenderAvatar = getString(bi, "uri")
					}
				} else {
					msg.SenderFbid = getString(sender, "id")
				}
			}

			// Fallback: lookup from participants
			if msg.SenderName == "" && msg.SenderFbid != "" {
				if u, ok := senderMap[msg.SenderFbid]; ok {
					msg.SenderName = u.FullName
					if msg.SenderAvatar == "" {
						msg.SenderAvatar = u.Avatar
					}
				}
			}

			// Attachments
			if msg.Text == "" {
				if blobs, ok := msgNode["blob_attachments"].([]interface{}); ok && len(blobs) > 0 {
					msg.Text = "[Đính kèm]"
					msg.ContentType = "attachment"
				}
				if sticker, _ := msgNode["sticker"].(map[string]interface{}); sticker != nil {
					msg.Text = "[Sticker]"
					msg.ContentType = "sticker"
				}
				if msg.Text == "" {
					if ext, _ := msgNode["extensible_attachment"].(map[string]interface{}); ext != nil {
						msg.Text = "[Đính kèm]"
						msg.ContentType = "attachment"
					}
				}
			}
			if msg.Text == "" {
				msg.Text = "[Đính kèm]"
				msg.ContentType = "attachment"
			}

			result.Messages = append(result.Messages, msg)
		}

		// Pagination: check page_info
		if pageInfo, _ := messages["page_info"].(map[string]interface{}); pageInfo != nil {
			if hasPrev, ok := pageInfo["has_previous_page"].(bool); ok && hasPrev {
				result.HasMore = true
				// Use oldest message timestamp as cursor
				if len(result.Messages) > 0 {
					oldest := result.Messages[len(result.Messages)-1]
					if oldest.TimestampMs != "" {
						ts, _ := strconv.ParseInt(oldest.TimestampMs, 10, 64)
						if ts > 0 {
							result.NextCursor = encodeGQLCursor(ts)
						}
					}
				}
			}
		}
	}

	// Fallback: last_message if no messages found
	if len(result.Messages) == 0 {
		if lastMsg, _ := msgThread["last_message"].(map[string]interface{}); lastMsg != nil {
			if lastNodes, _ := lastMsg["nodes"].([]interface{}); len(lastNodes) > 0 {
				if ln, _ := lastNodes[0].(map[string]interface{}); ln != nil {
					result.Messages = append(result.Messages, domain.ChatMessage{
						Text:        getString(ln, "snippet"),
						TimestampMs: getString(ln, "timestamp_precise"),
					})
				}
			}
		}
	}

	return result, nil
}

// ── Send Message (port of Python send_message) ──────────────────────────────

// SendMessageGQL sends a text message to a thread via GraphQL API.
// Port of Python: MessengerClient.send_message(text, thread_id)
func SendMessageGQL(cookieStr string, threadID string, text string) error {
	sess, err := getGQLSession(cookieStr)
	if err != nil {
		return err
	}

	otid := generateOfflineThreadingID()
	nowMs := time.Now().UnixMilli()

	// Build the send payload — same structure as Python label "46"
	sendPayload := map[string]interface{}{
		"thread_id":          threadID,
		"otid":               otid,
		"source":             0,
		"send_type":          1,
		"text":               text,
		"initiating_source":  1,
		"skip_url_preview_gen": 0,
		"sync_group":         1,
	}
	sendPayloadJSON, _ := json.Marshal(sendPayload)

	readPayload := map[string]interface{}{
		"thread_id":              threadID,
		"last_read_watermark_ts": nowMs,
		"sync_group":            1,
	}
	readPayloadJSON, _ := json.Marshal(readPayload)

	tasks := []map[string]interface{}{
		{
			"label":         "46",
			"payload":       string(sendPayloadJSON),
			"queue_name":    threadID,
			"task_id":       1,
			"failure_count": nil,
		},
		{
			"label":         "21",
			"payload":       string(readPayloadJSON),
			"queue_name":    threadID,
			"task_id":       2,
			"failure_count": nil,
		},
	}

	epochID := generateOfflineThreadingID()
	payloadObj := map[string]interface{}{
		"tasks":         tasks,
		"epoch_id":      epochID,
		"version_id":    "6120284488008082",
		"data_trace_id": nil,
	}
	payloadJSON, _ := json.Marshal(payloadObj)

	form := url.Values{
		"app_id":  {"2220391788200892"},
		"payload": {string(payloadJSON)},
		"type":    {"3"},
	}

	// Use /api/graphql/ endpoint for sending
	body, err := graphqlPost(sess, form)
	if err != nil {
		// Fallback: try the messaging send endpoint
		return sendViaMessagingEndpoint(sess, threadID, text)
	}

	respStr := string(body)
	if strings.Contains(respStr, "error") && strings.Contains(respStr, "Couldn't send") {
		return fmt.Errorf("facebook rejected the message")
	}

	slog.Info("gql message sent", "thread_id", threadID, "text_len", len(text))
	return nil
}

// sendViaMessagingEndpoint is a fallback using the /messaging/send/ endpoint
func sendViaMessagingEndpoint(sess *gqlSession, threadID string, text string) error {
	otid := generateOfflineThreadingID()
	nowMs := strconv.FormatInt(time.Now().UnixMilli(), 10)

	form := url.Values{
		"fb_dtsg":                     {sess.tokens.FbDtsg},
		"jazoest":                     {sess.tokens.Jazoest},
		"__user":                      {sess.tokens.FacebookID},
		"__a":                         {"1"},
		"body":                        {text},
		"other_user_fbid":             {threadID},
		"specific_to_list[0]":         {"fbid:" + threadID},
		"specific_to_list[1]":         {"fbid:" + sess.tokens.FacebookID},
		"has_attachment":              {"false"},
		"ephemeral_ttl_mode":          {"0"},
		"source":                      {"source:chat:web"},
		"client_thread_id":            {"root:" + otid},
		"offline_threading_id":        {otid},
		"message_id":                  {otid},
		"threading_id":               {"<" + nowMs + ":0-0@mail.projektitan.com>"},
		"timestamp":                   {nowMs},
	}

	req, _ := http.NewRequest("POST", fbBaseURL+"/messaging/send/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", sess.cookieStr)
	req.Header.Set("User-Agent", fbUA)
	req.Header.Set("Origin", fbBaseURL)
	req.Header.Set("Referer", fbBaseURL+"/messages/")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send message request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send message status %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	slog.Info("message sent via messaging endpoint", "thread_id", threadID)
	return nil
}

// ── Fetch Chat History (port of Python fetch_chat_history) ───────────────────

// FetchChatHistoryGQL fetches chat history for AI analysis.
// Returns list of messages in chronological order.
func FetchChatHistoryGQL(cookieStr string, threadID string, limit int) ([]map[string]interface{}, error) {
	threadResp, err := FetchThreadGQL(cookieStr, threadID, "")
	if err != nil {
		return nil, err
	}

	var history []map[string]interface{}
	// Messages are newest-first from API, reverse for chronological
	msgs := threadResp.Messages
	if len(msgs) > limit {
		msgs = msgs[:limit]
	}

	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		isMe := msg.SenderFbid == threadResp.ViewerID
		author := "Khách"
		if isMe {
			author = "Tôi"
		}
		ts, _ := strconv.ParseInt(msg.TimestampMs, 10, 64)
		history = append(history, map[string]interface{}{
			"author":    author,
			"text":      msg.Text,
			"timestamp": ts,
		})
	}

	return history, nil
}

// ── Cursor helpers ──────────────────────────────────────────────────────────

type gqlCursor struct {
	Before int64 `json:"before"`
}

func encodeGQLCursor(beforeTS int64) string {
	raw, _ := json.Marshal(gqlCursor{Before: beforeTS})
	return string(raw)
}

func decodeGQLCursor(cursor string) (*gqlCursor, error) {
	var c gqlCursor
	if err := json.Unmarshal([]byte(cursor), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// ── Utility functions (port of Python utils.py) ─────────────────────────────

// generateOfflineThreadingID generates an offline threading ID.
// Port of Python: generate_offline_threading_id()
func generateOfflineThreadingID() string {
	nowMs := time.Now().UnixMilli()
	value := rand.Int63n(4294967295)
	binary := fmt.Sprintf("%b", value)
	// Pad to 22 bits
	for len(binary) < 22 {
		binary = "0" + binary
	}
	binary = binary[len(binary)-22:]
	combined := fmt.Sprintf("%b", nowMs) + binary
	// Parse combined binary string
	result, _ := strconv.ParseInt(combined, 2, 64)
	return strconv.FormatInt(result, 10)
}
