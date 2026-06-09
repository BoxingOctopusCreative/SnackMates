package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultAPIURL = "http://localhost:8080"

type Client struct {
	baseURL string
	token   string
	userID  string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Token() string  { return c.token }
func (c *Client) UserID() string { return c.userID }

type apiError struct {
	status  int
	message string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%s (%d)", e.message, e.status)
}

type Me struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func (c *Client) Me() (*Me, error) {
	var me Me
	if err := c.get("/api/v1/auth/me", &me); err != nil {
		return nil, err
	}
	return &me, nil
}

func (c *Client) Login(email, password string) error {
	var resp struct {
		Token       string `json:"token"`
		MFARequired bool   `json:"mfa_required"`
	}
	if err := c.post("/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, &resp); err != nil {
		return err
	}
	if resp.MFARequired || resp.Token == "" {
		return fmt.Errorf("login requires MFA or returned no token")
	}
	c.token = resp.Token

	var me struct {
		ID string `json:"id"`
	}
	if err := c.get("/api/v1/auth/me", &me); err != nil {
		return err
	}
	c.userID = me.ID
	return nil
}

type Friendship struct {
	ID          string `json:"id"`
	RequesterID string `json:"requester_id"`
	AddresseeID string `json:"addressee_id"`
	Status      string `json:"status"`
	User        *struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	} `json:"user"`
}

type ChatMessage struct {
	ID        string `json:"id"`
	ChatID    string `json:"chat_id"`
	SenderID  string `json:"sender_id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type DirectMessage struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	Subject        string `json:"subject"`
	Body           string `json:"body"`
	CreatedAt      string `json:"created_at"`
}

type Conversation struct {
	ID          string `json:"id"`
	UnreadCount int    `json:"unread_count"`
	OtherUser   *struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	} `json:"other_user"`
}

type Chat struct {
	ID          string `json:"id"`
	UnreadCount int    `json:"unread_count"`
	OtherUser   *struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	} `json:"other_user"`
}

func (c *Client) FriendRequests() ([]Friendship, error) {
	var list []Friendship
	if err := c.get("/api/v1/friends/requests", &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *Client) RequestFriend(username string) (*Friendship, error) {
	var rec Friendship
	if err := c.post("/api/v1/friends/request", map[string]string{"username": username}, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (c *Client) AcceptFriend(friendshipID string) (*Friendship, error) {
	var rec Friendship
	if err := c.post("/api/v1/friends/"+friendshipID+"/accept", nil, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (c *Client) Chats() ([]Chat, int, error) {
	var resp struct {
		UnreadCount int    `json:"unread_count"`
		Chats       []Chat `json:"chats"`
	}
	if err := c.get("/api/v1/chats", &resp); err != nil {
		return nil, 0, err
	}
	return resp.Chats, resp.UnreadCount, nil
}

func (c *Client) GetChat(chatID string) (*Chat, []ChatMessage, error) {
	var resp struct {
		Chat     Chat          `json:"chat"`
		Messages []ChatMessage `json:"messages"`
	}
	if err := c.get("/api/v1/chats/"+chatID, &resp); err != nil {
		return nil, nil, err
	}
	return &resp.Chat, resp.Messages, nil
}

func (c *Client) StartChat(username string) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.post("/api/v1/chats", map[string]string{"username": username}, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) SendChatMessage(chatID, body string) (*ChatMessage, error) {
	var msg ChatMessage
	if err := c.post(
		"/api/v1/chats/"+chatID+"/messages",
		map[string]string{"body": body},
		&msg,
	); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) Conversations() ([]Conversation, int, error) {
	var resp struct {
		UnreadCount   int            `json:"unread_count"`
		Conversations []Conversation `json:"conversations"`
	}
	if err := c.get("/api/v1/messages/conversations", &resp); err != nil {
		return nil, 0, err
	}
	return resp.Conversations, resp.UnreadCount, nil
}

func (c *Client) GetConversation(conversationID string) (*Conversation, []DirectMessage, error) {
	var resp struct {
		Conversation Conversation    `json:"conversation"`
		Messages     []DirectMessage `json:"messages"`
	}
	if err := c.get("/api/v1/messages/conversations/"+conversationID, &resp); err != nil {
		return nil, nil, err
	}
	return &resp.Conversation, resp.Messages, nil
}

type SnagDelivery struct {
	Method         string
	TrackingNumber string
}

type WishlistSummary struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type SnaggedBy struct {
	ID string `json:"id"`
}

type WishlistItem struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	SnaggedBy *SnaggedBy `json:"snagged_by"`
}

type WishlistDetail struct {
	Items         []WishlistItem `json:"items"`
	ViewerCanSnag bool           `json:"viewer_can_snag"`
}

type UserProfile struct {
	User      PublicUser        `json:"user"`
	Wishlists []WishlistSummary `json:"wishlists"`
}

type PublicUser struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func (c *Client) Friends() ([]Friendship, error) {
	var list []Friendship
	if err := c.get("/api/v1/friends/", &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *Client) GetUserProfile(username string) (*UserProfile, error) {
	var profile UserProfile
	if err := c.get("/api/v1/users/"+username, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (c *Client) GetWishlist(slug string) (*WishlistDetail, error) {
	var detail WishlistDetail
	if err := c.get("/api/v1/wishlists/"+slug, &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

func (c *Client) SnagItem(wishlistSlug, itemID string, delivery SnagDelivery) (*WishlistItem, error) {
	method := delivery.Method
	if method == "" {
		method = "in_person"
	}
	body := map[string]string{"delivery_method": method}
	if method == "mail" && strings.TrimSpace(delivery.TrackingNumber) != "" {
		body["tracking_number"] = strings.TrimSpace(delivery.TrackingNumber)
	}

	var item WishlistItem
	if err := c.post(
		"/api/v1/wishlists/"+wishlistSlug+"/items/"+itemID+"/snag",
		body,
		&item,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (c *Client) SendDirectMessage(conversationID, subject, body string) (*DirectMessage, error) {
	var msg DirectMessage
	if err := c.post(
		"/api/v1/messages/conversations/"+conversationID+"/messages",
		map[string]string{"subject": subject, "body": body},
		&msg,
	); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) get(path string, out any) error {
	return c.request(http.MethodGet, path, nil, out)
}

func (c *Client) post(path string, body any, out any) error {
	return c.request(http.MethodPost, path, body, out)
}

func (c *Client) request(method, path string, body any, out any) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, r)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var errBody struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(raw, &errBody)
		msg := errBody.Error
		if msg == "" {
			msg = string(raw)
		}
		return &apiError{status: res.StatusCode, message: msg}
	}

	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}
