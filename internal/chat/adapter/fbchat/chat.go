package fbchat

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/chat/domain"
)

const (
	fbBaseURL = "https://www.facebook.com"
	fbUA      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
)

// fbTokens holds extracted tokens from the Facebook home page (ported from _session.py dataGetHome).
type fbTokens struct {
	FbDtsg         string
	Jazoest        string
	FacebookID     string // actorID
	ClientRevision string
}

// fetchFBTokens loads facebook.com with cookies and extracts auth tokens.
// Port of fbchat-v2 _session.py dataGetHome()
func fetchFBTokens(cookieStr string) (*fbTokens, error) {
	req, _ := http.NewRequest("GET", fbBaseURL+"/", nil)
	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("User-Agent", fbUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "vi-VN,vi;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fb home request: %w", err)
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	html := string(bodyBytes)

	tokens := &fbTokens{}

	// Extract fb_dtsg — same pattern as _session.py
	reDtsg := regexp.MustCompile(`DTSGInitialData",[^"]*\{"token":"([^"]+)"`)
	if m := reDtsg.FindStringSubmatch(html); len(m) > 1 {
		tokens.FbDtsg = m[1]
	}
	if tokens.FbDtsg == "" {
		// Fallback: try alternate pattern
		reDtsg2 := regexp.MustCompile(`"dtsg":\{"token":"([^"]+)"`)
		if m := reDtsg2.FindStringSubmatch(html); len(m) > 1 {
			tokens.FbDtsg = m[1]
		}
	}

	// Extract jazoest
	reJazoest := regexp.MustCompile(`jazoest=(\d+)`)
	if m := reJazoest.FindStringSubmatch(html); len(m) > 1 {
		tokens.Jazoest = m[1]
	}

	// Extract actorID (FacebookID)
	reActor := regexp.MustCompile(`"actorID":"(\d+)"`)
	if m := reActor.FindStringSubmatch(html); len(m) > 1 {
		tokens.FacebookID = m[1]
	}

	// Extract client_revision
	reRev := regexp.MustCompile(`"client_revision":(\d+)`)
	if m := reRev.FindStringSubmatch(html); len(m) > 1 {
		tokens.ClientRevision = m[1]
	}

	slog.Info("facebook tokens extracted",
		"dtsg_len", len(tokens.FbDtsg),
		"jazoest", tokens.Jazoest,
		"facebook_id", tokens.FacebookID,
		"client_revision", tokens.ClientRevision,
	)

	if tokens.FbDtsg == "" {
		return nil, fmt.Errorf("failed to extract fb_dtsg from facebook.com (cookie may be expired)")
	}

	return tokens, nil
}

// FetchInbox fetches Facebook Messenger inbox threads.
// Uses direct GraphQL API (ported from Python fbchat_muqit).
func FetchInbox(cookieStr string) (*domain.ChatInboxResponse, error) {
	// Use GraphQL batch API — same approach as Python fbchat_muqit
	result, err := FetchInboxGQL(cookieStr)
	if err == nil {
		return result, nil
	}
	slog.Warn("gql inbox failed, falling back to legacy graphql", "error", err)

	// Fallback to old GraphQL approach
	tokens, err := fetchFBTokens(cookieStr)
	if err != nil {
		return nil, err
	}

	// Build queries JSON — same structure as fbchat-v2
	queries := map[string]interface{}{
		"o0": map[string]interface{}{
			"doc_id": "3336396659757871",
			"query_params": map[string]interface{}{
				"limit":                  50,
				"before":                 nil,
				"tags":                   []string{"INBOX"},
				"includeDeliveryReceipts": false,
				"includeSeqID":           true,
			},
		},
	}
	queriesJSON, _ := json.Marshal(queries)

	form := url.Values{
		"fb_dtsg":  {tokens.FbDtsg},
		"jazoest":  {tokens.Jazoest},
		"__a":      {"1"},
		"__user":   {tokens.FacebookID},
		"av":       {tokens.FacebookID},
		"__rev":    {tokens.ClientRevision},
		"queries":  {string(queriesJSON)},
	}

	req, _ := http.NewRequest("POST", fbBaseURL+"/api/graphqlbatch/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("User-Agent", fbUA)
	req.Header.Set("Origin", fbBaseURL)
	req.Header.Set("Referer", fbBaseURL+"/messages/")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fb inbox request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	slog.Info("facebook graphql response",
		"status", resp.StatusCode,
		"body_len", len(body),
	)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fb inbox status %d: %s", resp.StatusCode, string(body[:min(len(body), 500)]))
	}

	// Parse GraphQL batch response — strip "for (;;);" prefix and parse multiple JSON objects
	text := string(body)
	if idx := strings.Index(text, "for (;;);"); idx >= 0 {
		text = text[idx+9:]
	}

	// Find the first JSON object with "o0" key
	parsed, err := parseGraphQLBatch(text)
	if err != nil {
		return nil, fmt.Errorf("parse fb response: %w (preview: %s)", err, text[:min(len(text), 300)])
	}

	return parseFBInbox(parsed, tokens.FacebookID)
}

// parseGraphQLBatch parses Facebook's GraphQL batch response (multiple JSON objects concatenated).
// Port of _all_thread_data.py _parse_graphqlbatch_response()
func parseGraphQLBatch(text string) (map[string]interface{}, error) {
	text = strings.TrimSpace(text)
	dec := json.NewDecoder(strings.NewReader(text))

	for dec.More() {
		var obj map[string]interface{}
		if err := dec.Decode(&obj); err != nil {
			return nil, err
		}
		if _, ok := obj["o0"]; ok {
			return obj, nil
		}
	}
	return nil, fmt.Errorf("no o0 object found in graphql batch response")
}

// parseFBInbox converts the raw GraphQL response into our domain model.
func parseFBInbox(data map[string]interface{}, viewerID string) (*domain.ChatInboxResponse, error) {
	// Navigate: o0 -> data -> viewer -> message_threads -> nodes
	o0, _ := data["o0"].(map[string]interface{})
	if o0 == nil {
		return nil, fmt.Errorf("missing o0 in response")
	}
	d, _ := o0["data"].(map[string]interface{})
	if d == nil {
		return nil, fmt.Errorf("missing data in o0")
	}
	viewer, _ := d["viewer"].(map[string]interface{})
	if viewer == nil {
		return nil, fmt.Errorf("missing viewer in data")
	}
	msgThreads, _ := viewer["message_threads"].(map[string]interface{})
	if msgThreads == nil {
		return nil, fmt.Errorf("missing message_threads")
	}
	nodes, _ := msgThreads["nodes"].([]interface{})

	result := &domain.ChatInboxResponse{
		ViewerID: viewerID,
	}

	for _, n := range nodes {
		node, ok := n.(map[string]interface{})
		if !ok {
			continue
		}

		thread := domain.ChatThread{}

		// Thread key
		threadKey, _ := node["thread_key"].(map[string]interface{})
		if threadKey != nil {
			if fbid, ok := threadKey["thread_fbid"].(string); ok && fbid != "" {
				thread.ThreadID = fbid
				thread.ThreadKey = fbid
			} else if otherID, ok := threadKey["other_user_id"].(string); ok && otherID != "" {
				thread.ThreadID = otherID
				thread.ThreadKey = otherID
			}
		}

		thread.Title = getString(node, "name")
		thread.IsGroup = thread.Title != "" // groups have names

		// Participants
		allParticipants, _ := node["all_participants"].(map[string]interface{})
		if allParticipants != nil {
			edges, _ := allParticipants["edges"].([]interface{})
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

				// Avatar
				bigImg, _ := actor["big_image_src"].(map[string]interface{})
				if bigImg != nil {
					user.Avatar = getString(bigImg, "uri")
				}

				thread.Users = append(thread.Users, user)
			}
		}

		// Set title from other participant for 1:1 chats
		if thread.Title == "" && len(thread.Users) >= 2 {
			for _, u := range thread.Users {
				if u.ID != viewerID {
					thread.Title = u.FullName
					if thread.Title == "" {
						thread.Title = u.Username
					}
					break
				}
			}
		}

		// Last message snippet
		messages, _ := node["last_message"].(map[string]interface{})
		if messages != nil {
			nodes2, _ := messages["nodes"].([]interface{})
			if len(nodes2) > 0 {
				msgNode, _ := nodes2[0].(map[string]interface{})
				if msgNode != nil {
					thread.LastMessage = getString(msgNode, "snippet")
				}
			}
		}

		// Timestamp
		if ts, ok := node["updated_time_precise"].(string); ok {
			thread.LastActivityMs = ts
		}

		// Messages count
		if mc, ok := node["messages_count"].(float64); ok {
			_ = mc
		}

		result.Threads = append(result.Threads, thread)
	}

	return result, nil
}

// FetchThread fetches messages for a specific Facebook Messenger thread.
// Uses direct GraphQL API (ported from Python fbchat_muqit).
func FetchThread(cookieStr string, threadID string, cursor string) (*domain.ChatThreadResponse, error) {
	result, err := FetchThreadGQL(cookieStr, threadID, cursor)
	if err == nil && result != nil {
		return result, nil
	}
	slog.Warn("gql thread failed, falling back to legacy graphql", "error", err)
	return fetchThreadGraphQL(cookieStr, threadID)
}

// SendMessage sends a text message to a Facebook Messenger thread.
func SendMessage(cookieStr string, threadID string, text string) error {
	return SendMessageGQL(cookieStr, threadID, text)
}

// fetchThreadGraphQL is the legacy GraphQL-based thread fetcher.
// Uses the same inbox query as fbchat-v2 (doc_id 3336396659757871) and filters by thread ID.
func fetchThreadGraphQL(cookieStr string, threadID string) (*domain.ChatThreadResponse, error) {
	tokens, err := fetchFBTokens(cookieStr)
	if err != nil {
		return nil, err
	}

	// Use the SAME inbox query that works
	queries := map[string]interface{}{
		"o0": map[string]interface{}{
			"doc_id": "3336396659757871",
			"query_params": map[string]interface{}{
				"limit":                   50,
				"before":                  nil,
				"tags":                    []string{"INBOX"},
				"includeDeliveryReceipts": false,
				"includeSeqID":            true,
			},
		},
	}
	queriesJSON, _ := json.Marshal(queries)

	form := url.Values{
		"fb_dtsg": {tokens.FbDtsg},
		"jazoest": {tokens.Jazoest},
		"__a":     {"1"},
		"__user":  {tokens.FacebookID},
		"av":      {tokens.FacebookID},
		"__rev":   {tokens.ClientRevision},
		"queries": {string(queriesJSON)},
	}

	req, _ := http.NewRequest("POST", fbBaseURL+"/api/graphqlbatch/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("User-Agent", fbUA)
	req.Header.Set("Origin", fbBaseURL)
	req.Header.Set("Referer", fbBaseURL+"/messages/")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fb thread request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fb thread status %d", resp.StatusCode)
	}

	text := string(body)
	if idx := strings.Index(text, "for (;;);"); idx >= 0 {
		text = text[idx+9:]
	}

	parsed, err := parseGraphQLBatch(text)
	if err != nil {
		return nil, fmt.Errorf("parse fb thread response: %w", err)
	}

	return findThreadInInbox(parsed, tokens.FacebookID, threadID)
}

// findThreadInInbox finds a specific thread by ID from the full inbox response.
func findThreadInInbox(data map[string]interface{}, viewerID, threadID string) (*domain.ChatThreadResponse, error) {
	o0, _ := data["o0"].(map[string]interface{})
	if o0 == nil {
		return nil, fmt.Errorf("missing o0")
	}
	d, _ := o0["data"].(map[string]interface{})
	if d == nil {
		return nil, fmt.Errorf("missing data")
	}
	viewer, _ := d["viewer"].(map[string]interface{})
	if viewer == nil {
		return nil, fmt.Errorf("missing viewer")
	}
	msgThreads, _ := viewer["message_threads"].(map[string]interface{})
	if msgThreads == nil {
		return nil, fmt.Errorf("missing message_threads")
	}
	nodes, _ := msgThreads["nodes"].([]interface{})

	result := &domain.ChatThreadResponse{
		ThreadID: threadID,
		ViewerID: viewerID,
	}

	for _, n := range nodes {
		node, ok := n.(map[string]interface{})
		if !ok {
			continue
		}
		threadKey, _ := node["thread_key"].(map[string]interface{})
		if threadKey == nil {
			continue
		}
		fbid := getString(threadKey, "thread_fbid")
		otherUID := getString(threadKey, "other_user_id")
		if fbid != threadID && otherUID != threadID {
			continue
		}

		// Found the thread
		result.Title = getString(node, "name")

		// Participants
		allP, _ := node["all_participants"].(map[string]interface{})
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

		// Title for 1:1
		if result.Title == "" && len(result.Users) >= 2 {
			for _, u := range result.Users {
				if u.ID != viewerID {
					result.Title = u.FullName
					break
				}
			}
		}

		// Messages
		messages, _ := node["messages"].(map[string]interface{})
		if messages != nil {
			msgNodes, _ := messages["nodes"].([]interface{})
			for _, mn := range msgNodes {
				msgNode, ok := mn.(map[string]interface{})
				if !ok {
					continue
				}
				msg := domain.ChatMessage{
					MessageID:   getString(msgNode, "message_id"),
					Text:        getString(msgNode, "snippet"),
					TimestampMs: getString(msgNode, "timestamp_precise"),
				}
				if sender, _ := msgNode["message_sender"].(map[string]interface{}); sender != nil {
					if actor, _ := sender["messaging_actor"].(map[string]interface{}); actor != nil {
						msg.SenderFbid = getString(actor, "id")
						msg.SenderName = getString(actor, "name")
						if bi, _ := actor["big_image_src"].(map[string]interface{}); bi != nil {
							msg.SenderAvatar = getString(bi, "uri")
						}
					}
				}
				if msg.Text == "" {
					msg.Text = getString(msgNode, "body")
				}
				if msg.Text == "" {
					if blobs, ok := msgNode["blob_attachments"].([]interface{}); ok && len(blobs) > 0 {
						msg.Text = "[Đính kèm]"
						msg.ContentType = "attachment"
					}
					if sticker, _ := msgNode["sticker"].(map[string]interface{}); sticker != nil {
						msg.Text = "[Sticker]"
						msg.ContentType = "sticker"
					}
				}
				result.Messages = append(result.Messages, msg)
			}
		}

		// Fallback: last_message
		if len(result.Messages) == 0 {
			if lastMsg, _ := node["last_message"].(map[string]interface{}); lastMsg != nil {
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

		break
	}

	return result, nil
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

