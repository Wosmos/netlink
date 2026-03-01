package auth

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	s := &AuthService{}

	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name@domain.co", true},
		{"user+tag@gmail.com", true},
		{"", false},
		{"notanemail", false},
		{"@domain.com", false},
		{"user@", false},
		{"user@.com", false},
	}

	for _, tc := range tests {
		if got := s.ValidateEmail(tc.email); got != tc.valid {
			t.Errorf("ValidateEmail(%q) = %v, want %v", tc.email, got, tc.valid)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	s := &AuthService{}

	tests := []struct {
		password string
		valid    bool
	}{
		{"12345678", true},
		{"a very long password that is within limits", true},
		{"short", false},
		{"1234567", false},
		{"", false},
	}

	for _, tc := range tests {
		if got := s.ValidatePassword(tc.password); got != tc.valid {
			t.Errorf("ValidatePassword(%q) = %v, want %v", tc.password, got, tc.valid)
		}
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	s := &AuthService{}
	password := "testpassword123"

	hash, err := s.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	if hash == password {
		t.Error("Hash should not equal plaintext password")
	}

	if !s.CheckPassword(password, hash) {
		t.Error("CheckPassword() should return true for correct password")
	}

	if s.CheckPassword("wrongpassword", hash) {
		t.Error("CheckPassword() should return false for wrong password")
	}
}

func TestGenerateToken(t *testing.T) {
	s := &AuthService{}

	token1, err := s.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	if len(token1) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("Token length = %d, want 32", len(token1))
	}

	token2, err := s.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	if token1 == token2 {
		t.Error("Two generated tokens should not be identical")
	}
}
