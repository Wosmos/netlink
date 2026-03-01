package websocket

import (
	"testing"
	"time"
)

func TestHubOnlineUsers(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Initially no users online
	if users := hub.GetOnlineUsers(); len(users) != 0 {
		t.Errorf("Expected 0 online users, got %d", len(users))
	}

	if hub.IsUserOnline(1) {
		t.Error("User 1 should not be online")
	}
}

func TestEventTypes(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventTypeMessage, "message"},
		{EventTypeTyping, "typing"},
		{EventTypeOnline, "online"},
		{EventTypeOffline, "offline"},
		{EventTypeRead, "read"},
		{EventTypeReaction, "reaction"},
		{EventTypeMessageDelete, "message_delete"},
		{EventTypeMessageEdit, "message_edit"},
	}

	for _, tc := range tests {
		if string(tc.eventType) != tc.expected {
			t.Errorf("EventType %v = %q, want %q", tc.eventType, string(tc.eventType), tc.expected)
		}
	}
}

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}
	if hub.clients == nil {
		t.Error("clients map should be initialized")
	}
	if hub.register == nil {
		t.Error("register channel should be initialized")
	}
	if hub.unregister == nil {
		t.Error("unregister channel should be initialized")
	}
	if hub.broadcast == nil {
		t.Error("broadcast channel should be initialized")
	}
}

func TestEventTimestamp(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	event := &Event{
		Type:   EventTypeMessage,
		UserID: 1,
	}

	hub.BroadcastToUsers([]int{999}, event)

	// Give the hub a moment to process
	time.Sleep(10 * time.Millisecond)

	if event.Timestamp.IsZero() {
		t.Error("BroadcastToUsers should set event timestamp")
	}
}
