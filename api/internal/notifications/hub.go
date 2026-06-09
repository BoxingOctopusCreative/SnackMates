package notifications

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/boxingoctopus/snackmates/api/internal/cache"
	"github.com/google/uuid"
)

const channel = "snackmates:notifications"

type Hub struct {
	mu     sync.Mutex
	nextID int
	subs   map[uuid.UUID]map[int]chan struct{}
	valkey *cache.Client
}

func NewHub(valkey *cache.Client) *Hub {
	return &Hub{
		subs:   make(map[uuid.UUID]map[int]chan struct{}),
		valkey: valkey,
	}
}

type pushMessage struct {
	UserID string `json:"user_id"`
}

func (h *Hub) Start(ctx context.Context) {
	if h.valkey == nil || !h.valkey.Available() {
		return
	}

	pubsub := h.valkey.Subscribe(ctx, channel)
	go func() {
		defer pubsub.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-pubsub.Channel():
				if !ok {
					return
				}
				var payload pushMessage
				if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
					continue
				}
				userID, err := uuid.Parse(payload.UserID)
				if err != nil {
					continue
				}
				h.notifyLocal(userID)
			}
		}
	}()
}

func (h *Hub) Subscribe(userID uuid.UUID) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)

	h.mu.Lock()
	if h.subs[userID] == nil {
		h.subs[userID] = make(map[int]chan struct{})
	}
	id := h.nextID
	h.nextID++
	h.subs[userID][id] = ch
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if userSubs, ok := h.subs[userID]; ok {
			delete(userSubs, id)
			if len(userSubs) == 0 {
				delete(h.subs, userID)
			}
		}
	}

	return ch, unsubscribe
}

func (h *Hub) Notify(userIDs ...uuid.UUID) {
	if len(userIDs) == 0 {
		return
	}

	seen := make(map[uuid.UUID]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID == uuid.Nil {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		h.notifyLocal(userID)
		h.publishRemote(userID)
	}
}

func (h *Hub) notifyLocal(userID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, ch := range h.subs[userID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (h *Hub) publishRemote(userID uuid.UUID) {
	if h.valkey == nil || !h.valkey.Available() {
		return
	}

	payload, err := json.Marshal(pushMessage{UserID: userID.String()})
	if err != nil {
		return
	}
	_ = h.valkey.Publish(context.Background(), channel, string(payload))
}
