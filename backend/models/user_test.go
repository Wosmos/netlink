package models

import (
	"testing"
	"time"
)

func TestUserIsOnline(t *testing.T) {
	// User with nil LastSeenAt
	u1 := &User{}
	if u1.IsOnline() {
		t.Error("User with nil LastSeenAt should not be online")
	}

	// User seen 1 minute ago — should be online
	recent := time.Now().Add(-1 * time.Minute)
	u2 := &User{LastSeenAt: &recent}
	if !u2.IsOnline() {
		t.Error("User seen 1 minute ago should be online")
	}

	// User seen 5 minutes ago — should be offline
	old := time.Now().Add(-5 * time.Minute)
	u3 := &User{LastSeenAt: &old}
	if u3.IsOnline() {
		t.Error("User seen 5 minutes ago should not be online")
	}
}

func TestMessageTypes(t *testing.T) {
	if MessageTypeText != "text" {
		t.Errorf("MessageTypeText = %q, want 'text'", MessageTypeText)
	}
	if MessageTypeVoice != "voice" {
		t.Errorf("MessageTypeVoice = %q, want 'voice'", MessageTypeVoice)
	}
}
