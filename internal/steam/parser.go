package steam

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) ParseItemURL(urlString string) (*ParsedItem, error) {

	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {

		return nil, fmt.Errorf("failed to create a request, err: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible)")

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch url, err: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam api returned status %d", resp.StatusCode)
	}

	item := &ParsedItem{Url: urlString}
	parsed, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsed.Host != "steamcommunity.com" && !strings.HasSuffix(parsed.Host, ".steamcommunity.com") {
		return nil, fmt.Errorf("unsupported host: %s", parsed.Host)
	}

	parts := strings.Split(parsed.Path, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid URL path: too short (%s)", parsed.Path)
	}

	if len(parts) < 4 || parts[1] != "market" || parts[2] != "listings" {
		return nil, fmt.Errorf("invalid Steam market URL format: expected /market/listings/{app_id}/")
	}

	appIDStr := parts[3]
	item.AppId, err = strconv.Atoi(appIDStr)
	if err != nil {
		return item, fmt.Errorf("invalid appID '%s': %w", appIDStr, err)
	}
	if item.AppId <= 0 {
		return nil, fmt.Errorf("invalid appID: must be positive")
	}

	hash := parts[4]
	item.HashName, err = url.QueryUnescape(hash)
	if err != nil {
		return item, fmt.Errorf("failed to unescape market hash name: %w", err)
	}
	if item.HashName == "" {
		return nil, fmt.Errorf("empty market hash name")
	}

	return item, nil
}
