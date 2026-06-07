// Package youtube provides a thin client for the YouTube Data API v3 search endpoint.
package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// ErrKeyNotConfigured is returned when no API key is set.
var ErrKeyNotConfigured = errors.New("youtube: YOUTUBE_API_KEY not configured")

const searchURL = "https://www.googleapis.com/youtube/v3/search"

// VideoResult holds the lean fields the frontend needs.
type VideoResult struct {
	VideoID      string `json:"videoId"`
	Title        string `json:"title"`
	ChannelName  string `json:"channelName"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

// Client calls the YouTube Data API v3 search endpoint.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient constructs a Client. Pass an empty string to get a no-op client that
// returns ErrKeyNotConfigured on every call.
func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, httpClient: &http.Client{}}
}

// Search queries YouTube for videos matching q and returns up to maxResults results.
func (c *Client) Search(ctx context.Context, q string, maxResults int) ([]VideoResult, error) {
	if c.apiKey == "" {
		return nil, ErrKeyNotConfigured
	}

	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("type", "video")
	params.Set("q", q)
	params.Set("maxResults", fmt.Sprintf("%d", maxResults))
	params.Set("key", c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("youtube.Search: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube.Search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("youtube.Search: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				ChannelTitle string `json:"channelTitle"`
				Thumbnails   struct {
					Medium struct {
						URL string `json:"url"`
					} `json:"medium"`
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
				} `json:"thumbnails"`
			} `json:"snippet"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("youtube.Search: %w", err)
	}

	results := make([]VideoResult, 0, len(body.Items))
	for _, item := range body.Items {
		thumb := item.Snippet.Thumbnails.Medium.URL
		if thumb == "" {
			thumb = item.Snippet.Thumbnails.Default.URL
		}
		results = append(results, VideoResult{
			VideoID:      item.ID.VideoID,
			Title:        item.Snippet.Title,
			ChannelName:  item.Snippet.ChannelTitle,
			ThumbnailURL: thumb,
		})
	}
	return results, nil
}
