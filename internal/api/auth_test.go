package api

import (
	"strings"
	"testing"
)

func TestBuildAuthorizeURL(t *testing.T) {
	url := buildAuthorizeURL("http://localhost:9999", "test-state")
	if url == "" {
		t.Error("expected non-empty URL")
	}
	if !strings.Contains(url, "login.microsoftonline.com") {
		t.Error("expected Microsoft login URL")
	}
	if !strings.Contains(url, "test-state") {
		t.Error("expected state parameter")
	}
	if !strings.Contains(url, "localhost") {
		t.Error("expected redirect URI")
	}
}

func TestRandomStateUniqueness(t *testing.T) {
	s1 := randomState()
	s2 := randomState()
	if s1 == s2 {
		t.Error("expected unique states")
	}
	if len(s1) != 32 {
		t.Errorf("expected 32 hex chars, got %d", len(s1))
	}
}
