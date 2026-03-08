package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

const baseURL = "https://api.powerbi.com/v1.0/myorg"
const fabricBaseURL = "https://api.fabric.microsoft.com"

// partitionTypeM is the TMDL partition type for M (Power Query) import tables.
const partitionTypeM = "m"

type refreshObject struct {
	Table string `json:"table"`
}

type refreshBody struct {
	Type       string          `json:"type"`
	CommitMode string          `json:"commitMode"`
	Objects    []refreshObject `json:"objects"`
}

func buildRefreshBody(tables []string) refreshBody {
	body := refreshBody{
		Type:       "full",
		CommitMode: "transactional",
	}
	if len(tables) > 0 {
		objects := make([]refreshObject, len(tables))
		for i, t := range tables {
			objects[i] = refreshObject{Table: t}
		}
		body.Objects = objects
	}
	return body
}

func parseRequestID(locationURL string) (string, error) {
	trimmed := strings.TrimRight(locationURL, "/")
	parts := strings.Split(trimmed, "/")
	id := parts[len(parts)-1]
	if err := validateUUID(id, "request ID"); err != nil {
		return "", fmt.Errorf("unexpected Location header %q: %w", locationURL, err)
	}
	return id, nil
}

func authHeader(token string) http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+token)
	return h
}

// maxResponseSize limits API response reads to 10 MB.
const maxResponseSize = 10 << 20

var httpClient = &http.Client{Timeout: 30 * time.Second}

func doGet(token, rawURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid request URL: %w", err)
	}
	req.Header = authHeader(token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// doPost sends an authenticated POST request and returns the response and body.
// The caller is responsible for checking resp.StatusCode for non-error codes like 202.
func doPost(token, rawURL string, reqBody io.Reader) (*http.Response, []byte, error) {
	req, err := http.NewRequest("POST", rawURL, reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid request URL: %w", err)
	}
	req.Header = authHeader(token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}
	return resp, body, nil
}

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func validateUUID(id, label string) error {
	if !uuidRe.MatchString(id) {
		return fmt.Errorf("invalid %s: %q is not a valid UUID", label, id)
	}
	return nil
}

func GetWorkspaceID(token, workspaceName string) (string, error) {
	filter := fmt.Sprintf("name eq '%s'", strings.ReplaceAll(workspaceName, "'", "''"))
	url := fmt.Sprintf("%s/groups?$filter=%s", baseURL, neturl.QueryEscape(filter))
	data, err := doGet(token, url)
	if err != nil {
		return "", err
	}
	var result struct {
		Value []struct {
			ID string `json:"id"`
		} `json:"value"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse workspace response: %w", err)
	}
	if len(result.Value) == 0 {
		return "", fmt.Errorf("workspace '%s' not found", workspaceName)
	}
	id := result.Value[0].ID
	if err := validateUUID(id, "workspace ID"); err != nil {
		return "", err
	}
	return id, nil
}

// Dataset represents a Power BI dataset (semantic model) in a workspace.
type Dataset struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListDatasets returns all datasets in a workspace.
func ListDatasets(token, workspaceID string) ([]Dataset, error) {
	if err := validateUUID(workspaceID, "workspace ID"); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/groups/%s/datasets", baseURL, workspaceID)
	data, err := doGet(token, url)
	if err != nil {
		return nil, err
	}
	var result struct {
		Value []Dataset `json:"value"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse dataset response: %w", err)
	}
	return result.Value, nil
}

// QueryTables returns the refreshable table names in a deployed semantic model.
// It uses the Fabric getDefinition API to fetch TMDL metadata, then includes
// only tables with an M (Power Query) partition — excluding calculated tables,
// calculation groups, and measure-only tables (which have no partition line).
func QueryTables(token, workspaceID, datasetID string) ([]string, error) {
	if err := validateUUID(workspaceID, "workspace ID"); err != nil {
		return nil, err
	}
	if err := validateUUID(datasetID, "dataset ID"); err != nil {
		return nil, err
	}

	parts, err := getDefinition(token, workspaceID, datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model definition: %w", err)
	}

	var tables []string
	for _, part := range parts {
		if !strings.Contains(part.Path, "/tables/") || !strings.HasSuffix(part.Path, ".tmdl") {
			continue
		}

		content := part.Payload
		// Fabric may return payloads as base64-encoded strings
		if decoded, decErr := base64.StdEncoding.DecodeString(content); decErr == nil {
			content = string(decoded)
		}

		tableName, partitionType := parseTMDL(content)
		if tableName != "" && partitionType == partitionTypeM {
			tables = append(tables, tableName)
		}
	}

	sort.Strings(tables)
	return tables, nil
}

type definitionPart struct {
	Path    string `json:"path"`
	Payload string `json:"payload"`
}

// getDefinition fetches the TMDL definition of a semantic model via the Fabric API.
// Handles the async 202 pattern: POST → poll operation → GET result.
func getDefinition(token, workspaceID, datasetID string) ([]definitionPart, error) {
	url := fmt.Sprintf("%s/v1/workspaces/%s/semanticModels/%s/getDefinition", fabricBaseURL, workspaceID, datasetID)
	resp, body, err := doPost(token, url, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("getDefinition API error %d: %s", resp.StatusCode, string(body))
	}

	// 202 = async operation, need to poll
	if resp.StatusCode == 202 {
		location := resp.Header.Get("Location")
		if location == "" {
			return nil, fmt.Errorf("getDefinition returned 202 but no Location header")
		}

		body, err = pollOperation(token, location)
		if err != nil {
			return nil, err
		}
	}

	var defResp struct {
		Definition struct {
			Parts []definitionPart `json:"parts"`
		} `json:"definition"`
	}
	if err := json.Unmarshal(body, &defResp); err != nil {
		return nil, fmt.Errorf("failed to parse definition response: %w", err)
	}
	return defResp.Definition.Parts, nil
}

// pollOperation polls a Fabric long-running operation until it succeeds,
// then fetches and returns the result body.
func pollOperation(token, operationURL string) ([]byte, error) {
	const maxAttempts = 30
	const pollInterval = 1 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		data, err := doGet(token, operationURL)
		if err != nil {
			return nil, fmt.Errorf("operation poll failed: %w", err)
		}

		var op struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &op); err != nil {
			return nil, fmt.Errorf("failed to parse operation status: %w", err)
		}

		switch op.Status {
		case "Succeeded":
			return doGet(token, operationURL+"/result")
		case "Failed":
			return nil, fmt.Errorf("operation failed: %s", string(data))
		}
		// Running or other — wait before next poll
		time.Sleep(pollInterval)
	}
	return nil, fmt.Errorf("operation did not complete within %d attempts", maxAttempts)
}

// parseTMDL extracts the table name and partition type from TMDL content in a single pass.
// Returns empty strings for files that don't contain a table definition.
func parseTMDL(content string) (tableName string, partitionType string) {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if tableName == "" {
			if strings.HasPrefix(trimmed, "///") {
				continue
			}
			if strings.HasPrefix(trimmed, "table ") {
				tableName = strings.Trim(strings.TrimPrefix(trimmed, "table "), "'")
			}
		}
		if strings.HasPrefix(trimmed, "partition ") {
			parts := strings.Split(trimmed, " = ")
			if len(parts) >= 2 {
				partitionType = strings.TrimSpace(parts[len(parts)-1])
				return // both found, bail early
			}
		}
	}
	return
}

func TriggerRefresh(token, workspaceID, datasetID string, tables []string) (string, error) {
	if err := validateUUID(workspaceID, "workspace ID"); err != nil {
		return "", err
	}
	if err := validateUUID(datasetID, "dataset ID"); err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(buildRefreshBody(tables))
	if err != nil {
		return "", fmt.Errorf("failed to marshal refresh body: %w", err)
	}

	url := fmt.Sprintf("%s/groups/%s/datasets/%s/refreshes", baseURL, workspaceID, datasetID)
	resp, body, err := doPost(token, url, bytes.NewReader(jsonData))
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("refresh API error %d: %s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header in refresh API response")
	}
	return parseRequestID(location)
}

type RefreshStatus struct {
	Status   string           `json:"status"`
	Messages []RefreshMessage `json:"messages,omitempty"`
}

type RefreshMessage struct {
	Message string `json:"message"`
}

func PollRefreshStatus(token, workspaceID, datasetID, requestID string) (RefreshStatus, error) {
	if err := validateUUID(workspaceID, "workspace ID"); err != nil {
		return RefreshStatus{}, err
	}
	if err := validateUUID(datasetID, "dataset ID"); err != nil {
		return RefreshStatus{}, err
	}
	url := fmt.Sprintf("%s/groups/%s/datasets/%s/refreshes/%s", baseURL, workspaceID, datasetID, requestID)
	data, err := doGet(token, url)
	if err != nil {
		return RefreshStatus{}, err
	}
	var status RefreshStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return RefreshStatus{}, fmt.Errorf("failed to parse refresh status: %w", err)
	}
	return status, nil
}

func WaitForRefresh(token, workspaceID, datasetID, requestID string, pollInterval, timeout time.Duration) (RefreshStatus, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := PollRefreshStatus(token, workspaceID, datasetID, requestID)
		if err != nil {
			return status, err
		}
		switch status.Status {
		case "Completed", "Failed", "Cancelled", "Disabled":
			return status, nil
		}
		time.Sleep(pollInterval)
	}
	return RefreshStatus{}, fmt.Errorf("refresh did not complete within %s", timeout)
}
