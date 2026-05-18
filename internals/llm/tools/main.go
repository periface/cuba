package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/periface/cuba/internals/utils"
	"github.com/serpapi/serpapi-golang"
)

type SerpAPISearchTool struct {
	client *serpapi.SerpApiClient
}

func NewSerpAPISearchTool() (*SerpAPISearchTool, error) {

	apiKey, err := utils.GetEnvVariable("SERPAPI_KEY")
	if err != nil {
		return nil, fmt.Errorf("missing SERPAPI_KEY: %w", err)
	}

	setting := serpapi.NewSerpApiClientSetting(apiKey)
	setting.Engine = "google"

	client := serpapi.NewClient(setting)

	return &SerpAPISearchTool{
		client: &client,
	}, nil
}

func (t *SerpAPISearchTool) Name() string {
	return "SerpAPI Google Search"
}

func (t *SerpAPISearchTool) Description() string {
	return `
Web search tool powered by SerpAPI (Google results).

Use it for:
- Companies
- Government corruption
- News
- Legal cases
- Contracts
- People / public figures

Input: plain text search query
`
}

func (t *SerpAPISearchTool) Call(ctx context.Context, input string) (string, error) {

	query := strings.TrimSpace(input)
	if query == "" {
		return "", fmt.Errorf("empty query")
	}

	if len(query) > 300 {
		query = query[:300]
	}

	// timeout control
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	params := map[string]string{
		"q":        query,
		"location": "Mexico",
		"hl":       "es",
		"gl":       "mx",
	}

	results, err := t.client.Search(params)
	if err != nil {
		return "", fmt.Errorf("serpapi search failed: %w", err)
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("SerpAPI results for: %s\n\n", query))

	organic, ok := results["organic_results"].([]interface{})
	if !ok || len(organic) == 0 {
		return "No results found", nil
	}

	for i, r := range organic {
		if i >= 8 {
			break
		}

		item, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		title, _ := item["title"].(string)
		link, _ := item["link"].(string)
		snippet, _ := item["snippet"].(string)

		if title == "" {
			title = "Untitled"
		}

		if snippet == "" {
			snippet = "No snippet"
		}

		sb.WriteString(fmt.Sprintf("Result %d\n", i+1))
		sb.WriteString(fmt.Sprintf("Title: %s\n", title))
		sb.WriteString(fmt.Sprintf("Snippet: %s\n", snippet))
		sb.WriteString(fmt.Sprintf("URL: %s\n", link))
		sb.WriteString("---\n")
	}

	return sb.String(), nil
}
