package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"time"
)

const baseURL = "https://api.powerbi.com/v1.0/myorg"

type refreshObject struct {
	Table string `json:"table"`
}

type refreshBody struct {
	Type       string          `json:"type"`
	CommitMode string          `json:"commitMode"`
	Objects    []refreshObject `json:"objects"`
}

func buildRefreshBody(tables []string, refreshType string) refreshBody {
	objects := make([]refreshObject, len(tables))
	for i, t := range tables {
		objects[i] = refreshObject{Table: t}
	}
	return refreshBody{
		Type:       refreshType,
		CommitMode: "transactional",
		Objects:    objects,
	}
}

func parseRequestID(locationURL string) string {
	trimmed := strings.TrimRight(locationURL, "/")
	parts := strings.Split(trimmed, "/")
	return parts[len(parts)-1]
}

func authHeader(token string) http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+token)
	return h
}

func doGet(token, url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = authHeader(token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
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
	return result.Value[0].ID, nil
}

func GetDatasetID(token, workspaceID, datasetName string) (string, error) {
	url := fmt.Sprintf("%s/groups/%s/datasets", baseURL, workspaceID)
	data, err := doGet(token, url)
	if err != nil {
		return "", err
	}
	var result struct {
		Value []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"value"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse dataset response: %w", err)
	}
	for _, ds := range result.Value {
		if ds.Name == datasetName {
			return ds.ID, nil
		}
	}
	return "", fmt.Errorf("dataset '%s' not found in workspace", datasetName)
}

func TriggerRefresh(token, workspaceID, datasetID string, tables []string) (string, error) {
	body := buildRefreshBody(tables, "full")
	jsonData, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/groups/%s/datasets/%s/refreshes", baseURL, workspaceID, datasetID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	req.Header = authHeader(token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("refresh API error %d: %s", resp.StatusCode, string(b))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header in refresh API response")
	}
	return parseRequestID(location), nil
}

type RefreshStatus struct {
	Status   string          `json:"status"`
	Messages []RefreshMessage `json:"messages,omitempty"`
}

type RefreshMessage struct {
	Message string `json:"message"`
}

func PollRefreshStatus(token, workspaceID, datasetID, requestID string) (RefreshStatus, error) {
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
