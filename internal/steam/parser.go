package steam

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) ParseItemURL(urlString string) (*ParsedItem, error) {
	item := &ParsedItem{Url: urlString}
	parsed, err := url.Parse(urlString)
	if err != nil {
		return item, fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsed.Host != "steamcommunity.com" && !strings.HasSuffix(parsed.Host, ".steamcommunity.com") {
		return item, fmt.Errorf("unsupported host: %s", parsed.Host)
	}

	parts := strings.Split(parsed.Path, "/")
	if len(parts) < 5 {
		return item, fmt.Errorf("invalid URL path: too short (%s)", parsed.Path)
	}

	if len(parts) < 4 || parts[1] != "market" || parts[2] != "listings" {
		return item, fmt.Errorf("invalid Steam market URL format: expected /market/listings/{app_id}/")
	}

	appIDStr := parts[3]
	item.AppId, err = strconv.Atoi(appIDStr)
	if err != nil {
		return item, fmt.Errorf("invalid appID '%s': %w", appIDStr, err)
	}
	if item.AppId <= 0 {
		return item, fmt.Errorf("invalid appID: must be positive")
	}

	hash := parts[4]
	item.HashName, err = url.QueryUnescape(hash)
	if err != nil {
		return item, fmt.Errorf("failed to unescape market hash name: %w", err)
	}
	if item.HashName == "" {
		return item, fmt.Errorf("empty market hash name")
	}

	return item, nil
}
