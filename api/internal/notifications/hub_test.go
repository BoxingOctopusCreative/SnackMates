package notifications

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHubNotifiesSubscriber(t *testing.T) {
	hub := NewHub(nil)
	userID := uuid.New()

	events, unsubscribe := hub.Subscribe(userID)
	defer unsubscribe()

	hub.Notify(userID)

	select {
	case <-events:
	case <-time.After(time.Second):
		t.Fatal("expected notification event")
	}
}

func TestHubUnsubscribeStopsDelivery(t *testing.T) {
	hub := NewHub(nil)
	userID := uuid.New()

	events, unsubscribe := hub.Subscribe(userID)
	unsubscribe()

	hub.Notify(userID)

	select {
	case <-events:
		t.Fatal("expected no event after unsubscribe")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHubDedupesDuplicateUserIDs(t *testing.T) {
	hub := NewHub(nil)
	userID := uuid.New()

	events, unsubscribe := hub.Subscribe(userID)
	defer unsubscribe()

	hub.Notify(userID, userID)

	select {
	case <-events:
	case <-time.After(time.Second):
		t.Fatal("expected notification event")
	}

	select {
	case <-events:
		t.Fatal("expected only one notification event")
	case <-time.After(50 * time.Millisecond):
	}
}
