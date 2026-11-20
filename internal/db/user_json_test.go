package db

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUserJSONDoesNotExposePasswordHash(t *testing.T) {
	encoded, err := json.Marshal(User{ID: 1, Name: "Ahmed", PasswordHash: "secret-hash"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secret-hash") || strings.Contains(string(encoded), "password_hash") {
		t.Fatalf("password hash leaked in user JSON: %s", encoded)
	}
}
