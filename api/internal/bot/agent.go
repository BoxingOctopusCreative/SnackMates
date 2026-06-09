package bot

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultChatReplyTemplate    = `Hi! I'm an automated test user. I got your chat: "{body}"`
	defaultMessageReplyTemplate = `Hi! I'm an automated test user. I got your message: "{body}"`
)

type Options struct {
	AcceptFriends      bool
	ReplyChats         bool
	ReplyMessages      bool
	SnagSnacks         bool
	SnagDeliveryMethod string
	SnagTrackingNumber string
	ReplyTemplate      string
	AddFriends         []string
	SendChats          []OutgoingMessage
	PollInterval       time.Duration
	DisplayName        string
}

type OutgoingMessage struct {
	Username string
	Body     string
}

type Agent struct {
	client *Client
	opts   Options

	mu        sync.Mutex
	repliedTo map[string]struct{}
}

func NewAgent(client *Client, opts Options) *Agent {
	if opts.PollInterval <= 0 {
		opts.PollInterval = 5 * time.Second
	}
	return &Agent{
		client:    client,
		opts:      opts,
		repliedTo: make(map[string]struct{}),
	}
}

func (a *Agent) Run(ctx context.Context) error {
	for _, username := range a.opts.AddFriends {
		if err := a.SendFriendRequest(username); err != nil {
			log.Printf("add-friend %s: %v", username, err)
		}
	}
	for _, msg := range a.opts.SendChats {
		if err := a.SendChatToUser(msg.Username, msg.Body); err != nil {
			log.Printf("send-chat %s: %v", msg.Username, err)
		}
	}

	a.tick()

	poll := time.NewTicker(a.opts.PollInterval)
	defer poll.Stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.stream(ctx, a.tick)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err != nil && ctx.Err() == nil {
				log.Printf("notification stream ended (%v); falling back to polling", err)
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-poll.C:
						a.tick()
					}
				}
			}
			return err
		case <-poll.C:
			a.tick()
		}
	}
}

func (a *Agent) SendFriendRequest(username string) error {
	rec, err := a.client.RequestFriend(username)
	if err != nil {
		return err
	}
	log.Printf("sent snack mate request to %s (friendship %s)", username, rec.ID)
	return nil
}

func (a *Agent) snagDelivery() SnagDelivery {
	return SnagDelivery{
		Method:         a.opts.SnagDeliveryMethod,
		TrackingNumber: a.opts.SnagTrackingNumber,
	}
}

func (a *Agent) SnagWishlistItem(wishlistSlug, itemID string) error {
	item, err := a.client.SnagItem(wishlistSlug, itemID, a.snagDelivery())
	if err != nil {
		return err
	}
	log.Printf("snagged %q on wishlist %s (item %s)", item.Name, wishlistSlug, item.ID)
	return nil
}

func (a *Agent) snagSnacks() {
	mates, err := a.client.Friends()
	if err != nil {
		log.Printf("list snack mates: %v", err)
		return
	}

	delivery := a.snagDelivery()
	if delivery.Method == "" {
		delivery.Method = "in_person"
	}

	for _, mate := range mates {
		if mate.User == nil || mate.User.Username == "" {
			continue
		}
		username := mate.User.Username
		profile, err := a.client.GetUserProfile(username)
		if err != nil {
			log.Printf("load profile %s: %v", username, err)
			continue
		}

		for _, wishlist := range profile.Wishlists {
			detail, err := a.client.GetWishlist(wishlist.Slug)
			if err != nil {
				log.Printf("load wishlist %s: %v", wishlist.Slug, err)
				continue
			}
			if !detail.ViewerCanSnag {
				continue
			}

			for _, item := range detail.Items {
				if item.SnaggedBy != nil {
					continue
				}
				snagged, err := a.client.SnagItem(wishlist.Slug, item.ID, delivery)
				if err != nil {
					log.Printf("snag %q on %s for %s: %v", item.Name, wishlist.Slug, username, err)
					continue
				}
				log.Printf("snagged %q on %s for %s (item %s)", snagged.Name, wishlist.Slug, username, snagged.ID)
			}
		}
	}
}

func (a *Agent) SendChatToUser(username, body string) error {
	chatID, err := a.client.StartChat(username)
	if err != nil {
		return err
	}
	msg, err := a.client.SendChatMessage(chatID, body)
	if err != nil {
		return err
	}
	log.Printf("sent chat to %s (chat %s, message %s)", username, chatID, msg.ID)
	return nil
}

func (a *Agent) tick() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.opts.AcceptFriends {
		a.acceptFriendRequests()
	}
	if a.opts.ReplyChats {
		a.replyToChats()
	}
	if a.opts.ReplyMessages {
		a.replyToMessages()
	}
	if a.opts.SnagSnacks {
		a.snagSnacks()
	}
}

func (a *Agent) acceptFriendRequests() {
	requests, err := a.client.FriendRequests()
	if err != nil {
		log.Printf("list friend requests: %v", err)
		return
	}
	for _, req := range requests {
		from := "unknown"
		if req.User != nil {
			from = req.User.Username
		}
		rec, err := a.client.AcceptFriend(req.ID)
		if err != nil {
			log.Printf("accept request from %s: %v", from, err)
			continue
		}
		log.Printf("accepted snack mate request from %s (friendship %s)", from, rec.ID)
	}
}

func (a *Agent) replyToChats() {
	chats, _, err := a.client.Chats()
	if err != nil {
		log.Printf("list chats: %v", err)
		return
	}

	for _, chat := range chats {
		if chat.UnreadCount == 0 {
			continue
		}
		_, messages, err := a.client.GetChat(chat.ID)
		if err != nil {
			log.Printf("load chat %s: %v", chat.ID, err)
			continue
		}

		latest := latestIncomingChatMessage(messages, a.client.UserID())
		if latest == nil {
			continue
		}
		if _, seen := a.repliedTo[latest.ID]; seen {
			continue
		}

		from := "friend"
		if chat.OtherUser != nil {
			from = chat.OtherUser.Username
		}
		reply := formatReply(a.chatReplyTemplate(), a.opts.DisplayName, from, latest.Body)
		if _, err := a.client.SendChatMessage(chat.ID, reply); err != nil {
			log.Printf("reply in chat with %s: %v", from, err)
			continue
		}
		a.repliedTo[latest.ID] = struct{}{}
		log.Printf("replied to %s in chat %s", from, chat.ID)
	}
}

func (a *Agent) replyToMessages() {
	conversations, _, err := a.client.Conversations()
	if err != nil {
		log.Printf("list conversations: %v", err)
		return
	}

	for _, conv := range conversations {
		if conv.UnreadCount == 0 {
			continue
		}
		_, messages, err := a.client.GetConversation(conv.ID)
		if err != nil {
			log.Printf("load conversation %s: %v", conv.ID, err)
			continue
		}

		latest := latestIncomingDirectMessage(messages, a.client.UserID())
		if latest == nil {
			continue
		}
		if _, seen := a.repliedTo[latest.ID]; seen {
			continue
		}

		from := "friend"
		if conv.OtherUser != nil {
			from = conv.OtherUser.Username
		}
		subject := replySubject(latest.Subject)
		reply := formatReply(a.messageReplyTemplate(), a.opts.DisplayName, from, latest.Body)
		if _, err := a.client.SendDirectMessage(conv.ID, subject, reply); err != nil {
			log.Printf("reply in conversation with %s: %v", from, err)
			continue
		}
		a.repliedTo[latest.ID] = struct{}{}
		log.Printf("replied to %s in conversation %s", from, conv.ID)
	}
}

func (a *Agent) chatReplyTemplate() string {
	if a.opts.ReplyTemplate != "" {
		return a.opts.ReplyTemplate
	}
	return defaultChatReplyTemplate
}

func (a *Agent) messageReplyTemplate() string {
	if a.opts.ReplyTemplate != "" {
		return a.opts.ReplyTemplate
	}
	return defaultMessageReplyTemplate
}

func replySubject(original string) string {
	if original == "" {
		return "Re: (no subject)"
	}
	if strings.HasPrefix(strings.ToLower(original), "re:") {
		return original
	}
	return "Re: " + original
}

func latestIncomingDirectMessage(messages []DirectMessage, selfID string) *DirectMessage {
	var latest *DirectMessage
	for i := range messages {
		msg := &messages[i]
		if msg.SenderID == selfID {
			continue
		}
		latest = msg
	}
	return latest
}

func latestIncomingChatMessage(messages []ChatMessage, selfID string) *ChatMessage {
	var latest *ChatMessage
	for i := range messages {
		msg := &messages[i]
		if msg.SenderID == selfID {
			continue
		}
		latest = msg
	}
	return latest
}

func formatReply(template, displayName, from, body string) string {
	replacer := strings.NewReplacer(
		"{display_name}", displayName,
		"{from}", from,
		"{body}", body,
	)
	return replacer.Replace(template)
}

func (a *Agent) stream(ctx context.Context, onRefresh func()) error {
	url := fmt.Sprintf("%s/api/v1/notifications/stream?access_token=%s",
		a.client.baseURL, a.client.Token())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("stream status %d", res.StatusCode)
	}

	scanner := bufio.NewScanner(res.Body)
	var event string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			event = strings.TrimPrefix(line, "event: ")
			continue
		}
		if line == "" {
			if event == "refresh" {
				onRefresh()
			}
			event = ""
		}
	}
	return scanner.Err()
}

func ParseOutgoingMessage(raw string) (OutgoingMessage, error) {
	idx := strings.Index(raw, ":")
	if idx <= 0 {
		return OutgoingMessage{}, fmt.Errorf("expected username:message body, got %q", raw)
	}
	username := strings.TrimSpace(raw[:idx])
	body := strings.TrimSpace(raw[idx+1:])
	if username == "" || body == "" {
		return OutgoingMessage{}, fmt.Errorf("username and message body required")
	}
	return OutgoingMessage{Username: username, Body: body}, nil
}
