package jellyfin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client represents a Jellyfin API client
type Client struct {
	ServerURL string
	APIKey    string
	HTTPClient *http.Client
}

// NewClient creates a new Jellyfin client
func NewClient(serverURL, apiKey string) *Client {
	return &Client{
		ServerURL: serverURL,
		APIKey:    apiKey,
		HTTPClient: &http.Client{},
	}
}

// MediaItem represents a movie, TV show, or episode
type MediaItem struct {
	ID           string            `json:"Id"`
	Name         string            `json:"Name"`
	Type         string            `json:"Type"`
	MediaType    string            `json:"MediaType"`
	ImageTags    map[string]string `json:"ImageTags"`
	IndexNumber  int               `json:"IndexNumber"`
}

// GetMovies fetches movies from the Jellyfin server
func (c *Client) GetMovies() ([]MediaItem, error) {
	endpoint := fmt.Sprintf("%s/Items?IncludeItemTypes=Movie&Recursive=true&api_key=%s", 
		c.ServerURL, c.APIKey)
	
	return c.fetchItems(endpoint)
}

// GetTVShows fetches TV shows from the Jellyfin server
func (c *Client) GetTVShows() ([]MediaItem, error) {
	endpoint := fmt.Sprintf("%s/Items?IncludeItemTypes=Series&Recursive=true&api_key=%s", 
		c.ServerURL, c.APIKey)
	
	return c.fetchItems(endpoint)
}

// Search searches for media items
func (c *Client) Search(query string) ([]MediaItem, error) {
	endpoint := fmt.Sprintf("%s/Items?SearchTerm=%s&Recursive=true&api_key=%s", 
		c.ServerURL, url.QueryEscape(query), c.APIKey)
	
	return c.fetchItems(endpoint)
}

// GetStreamURL returns the streaming URL for a media item
func (c *Client) GetStreamURL(itemID string) string {
	return fmt.Sprintf("%s/Videos/%s/stream?api_key=%s", c.ServerURL, itemID, c.APIKey)
}

// Helper function to fetch items from an endpoint
func (c *Client) fetchItems(endpoint string) ([]MediaItem, error) {
	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var response struct {
		Items []MediaItem `json:"Items"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	
	return response.Items, nil
}

// FetchItems fetches items from a custom endpoint
func (c *Client) FetchItems(endpoint string) ([]MediaItem, error) {
	return c.fetchItems(endpoint)
} 