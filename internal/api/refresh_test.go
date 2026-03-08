package api

import (
	"testing"
)

func TestBuildRefreshBody(t *testing.T) {
	body := buildRefreshBody([]string{"Dim Employee", "Fact Absence"})
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

func TestBuildRefreshBodyFullModel(t *testing.T) {
	body := buildRefreshBody(nil)
	if body.Type != "full" {
		t.Errorf("expected full, got %s", body.Type)
	}
	if body.Objects != nil {
		t.Errorf("expected nil objects for full model refresh, got %v", body.Objects)
	}
}

func TestParseRequestID(t *testing.T) {
	url := "https://api.powerbi.com/v1.0/myorg/groups/abc/datasets/def/refreshes/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	id, err := parseRequestID(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Errorf("expected aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee, got %s", id)
	}
}

func TestParseRequestIDTrailingSlash(t *testing.T) {
	url := "https://api.powerbi.com/v1.0/myorg/groups/abc/datasets/def/refreshes/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee/"
	id, err := parseRequestID(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Errorf("expected aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee, got %s", id)
	}
}

func TestParseRequestIDInvalid(t *testing.T) {
	url := "https://api.powerbi.com/v1.0/myorg/groups/abc/datasets/def/refreshes/not-a-uuid"
	_, err := parseRequestID(url)
	if err == nil {
		t.Error("expected error for invalid request ID, got nil")
	}
}

func TestParseTMDL(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		wantTable     string
		wantPartition string
	}{
		{
			name: "import table",
			content: `/// Source tables (Silver):
/// - SomeSource.Table
table 'Dim Employee'
	lineageTag: abc-123

	partition 'Dim Employee' = m
		mode: import`,
			wantTable:     "Dim Employee",
			wantPartition: "m",
		},
		{
			name: "calculated table",
			content: `table _MeasuresModell
	lineageTag: abc-123

	partition _MeasuresModell = calculated
		mode: import`,
			wantTable:     "_MeasuresModell",
			wantPartition: "calculated",
		},
		{
			name: "no partition (measure-only)",
			content: `table _MeasuresReport
	lineageTag: abc-123

	column dummy
		isHidden
		summarizeBy: none`,
			wantTable:     "_MeasuresReport",
			wantPartition: "",
		},
		{
			name:          "empty content",
			content:       "",
			wantTable:     "",
			wantPartition: "",
		},
		{
			name: "table without quotes",
			content: `table SomeTable
	partition SomeTable = m
		mode: import`,
			wantTable:     "SomeTable",
			wantPartition: "m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, partitionType := parseTMDL(tt.content)
			if tableName != tt.wantTable {
				t.Errorf("tableName = %q, want %q", tableName, tt.wantTable)
			}
			if partitionType != tt.wantPartition {
				t.Errorf("partitionType = %q, want %q", partitionType, tt.wantPartition)
			}
		})
	}
}
