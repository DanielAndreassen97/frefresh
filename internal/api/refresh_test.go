package api

import (
	"testing"
)

func TestBuildRefreshBody(t *testing.T) {
	body := buildRefreshBody([]string{"Dim Employee", "Fact Absence"}, "full")
	if body.Type != "full" {
		t.Errorf("expected full, got %s", body.Type)
	}
	if len(body.Objects) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(body.Objects))
	}
	if body.Objects[0].Table != "Dim Employee" {
		t.Errorf("expected Dim Employee, got %s", body.Objects[0].Table)
	}
	if body.CommitMode != "transactional" {
		t.Errorf("expected transactional, got %s", body.CommitMode)
	}
}

func TestParseRequestID(t *testing.T) {
	url := "https://api.powerbi.com/v1.0/myorg/groups/abc/datasets/def/refreshes/12345"
	id := parseRequestID(url)
	if id != "12345" {
		t.Errorf("expected 12345, got %s", id)
	}
}

func TestParseRequestIDTrailingSlash(t *testing.T) {
	url := "https://api.powerbi.com/v1.0/myorg/groups/abc/datasets/def/refreshes/12345/"
	id := parseRequestID(url)
	if id != "12345" {
		t.Errorf("expected 12345, got %s", id)
	}
}
