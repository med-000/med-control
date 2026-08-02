package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	apiVersion = "2026-03-11"
	baseURL    = "https://api.notion.com"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (client *Client) RetrieveDatabase(ctx context.Context, databaseID string) (map[string]any, error) {
	endpoint := fmt.Sprintf("%s/v1/databases/%s", baseURL, url.PathEscape(databaseID))

	return client.doJSON(ctx, http.MethodGet, endpoint, nil)
}

func (client *Client) RetrieveDatabaseRows(ctx context.Context, databaseID string) ([]map[string]any, error) {
	database, err := client.RetrieveDatabase(ctx, databaseID)
	if err != nil {
		return nil, err
	}

	dataSourceID, err := firstDataSourceID(database)
	if err != nil {
		return nil, err
	}

	return client.QueryDataSource(ctx, dataSourceID)
}

func (client *Client) RetrieveDatabaseTasks(ctx context.Context, databaseID string) ([]map[string]any, error) {
	rows, err := client.RetrieveDatabaseRows(ctx, databaseID)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		pageID, ok := row["id"].(string)
		if !ok || pageID == "" {
			continue
		}

		description, err := client.RetrievePageMarkdown(ctx, pageID)
		if err != nil {
			return nil, err
		}
		row["description"] = description
	}

	return rows, nil
}

func (client *Client) QueryDataSource(ctx context.Context, dataSourceID string) ([]map[string]any, error) {
	var rows []map[string]any
	var startCursor string

	for {
		requestBody := map[string]any{
			"page_size": 100,
		}
		if startCursor != "" {
			requestBody["start_cursor"] = startCursor
		}

		endpoint := fmt.Sprintf("%s/v1/data_sources/%s/query", baseURL, url.PathEscape(dataSourceID))
		response, err := client.doJSON(ctx, http.MethodPost, endpoint, requestBody)
		if err != nil {
			return nil, err
		}

		results, ok := response["results"].([]any)
		if !ok {
			return nil, fmt.Errorf("notion query data source response does not include results")
		}

		for _, result := range results {
			row, ok := result.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("notion query data source result has unexpected shape")
			}
			rows = append(rows, row)
		}

		hasMore, _ := response["has_more"].(bool)
		if !hasMore {
			return rows, nil
		}

		nextCursor, _ := response["next_cursor"].(string)
		if nextCursor == "" {
			return nil, fmt.Errorf("notion query data source has_more is true but next_cursor is empty")
		}
		startCursor = nextCursor
	}
}

func (client *Client) RetrievePageMarkdown(ctx context.Context, pageID string) (string, error) {
	blocks, err := client.retrieveBlockChildren(ctx, pageID)
	if err != nil {
		return "", err
	}

	return blocksToMarkdown(blocks, 0), nil
}

func (client *Client) retrieveBlockChildren(ctx context.Context, blockID string) ([]map[string]any, error) {
	var blocks []map[string]any
	var startCursor string

	for {
		endpoint := fmt.Sprintf("%s/v1/blocks/%s/children?page_size=100", baseURL, url.PathEscape(blockID))
		if startCursor != "" {
			endpoint += "&start_cursor=" + url.QueryEscape(startCursor)
		}

		response, err := client.doJSON(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}

		results, ok := response["results"].([]any)
		if !ok {
			return nil, fmt.Errorf("notion retrieve block children response does not include results")
		}

		for _, result := range results {
			block, ok := result.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("notion block child has unexpected shape")
			}

			hasChildren, _ := block["has_children"].(bool)
			if hasChildren {
				children, err := client.retrieveBlockChildren(ctx, stringValue(block["id"]))
				if err != nil {
					return nil, err
				}
				block["children"] = children
			}

			blocks = append(blocks, block)
		}

		hasMore, _ := response["has_more"].(bool)
		if !hasMore {
			return blocks, nil
		}

		nextCursor, _ := response["next_cursor"].(string)
		if nextCursor == "" {
			return nil, fmt.Errorf("notion retrieve block children has_more is true but next_cursor is empty")
		}
		startCursor = nextCursor
	}
}

func (client *Client) doJSON(ctx context.Context, method string, endpoint string, requestBody map[string]any) (map[string]any, error) {
	var bodyReader io.Reader
	if requestBody != nil {
		body, err := json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(body)
	}

	request, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+client.apiKey)
	request.Header.Set("Notion-Version", apiVersion)
	request.Header.Set("Accept", "application/json")
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("notion request failed: method=%s endpoint=%s status=%d body=%s", method, endpoint, response.StatusCode, string(body))
	}

	var responseBody map[string]any
	if err := json.Unmarshal(body, &responseBody); err != nil {
		return nil, err
	}

	return responseBody, nil
}

func firstDataSourceID(database map[string]any) (string, error) {
	dataSources, ok := database["data_sources"].([]any)
	if !ok || len(dataSources) == 0 {
		return "", fmt.Errorf("notion database response does not include data_sources")
	}

	firstDataSource, ok := dataSources[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("notion database data_sources has unexpected shape")
	}

	id, ok := firstDataSource["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("notion database data source does not include id")
	}

	return id, nil
}
