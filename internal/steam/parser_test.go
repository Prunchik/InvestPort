package steam

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseItemURL_ValidURLs(t *testing.T) {
	client := &Client{}

	test := []struct {
		name     string
		url      string
		expected *ParsedItem
	}{
		{
			name: "basic market listing",
			url:  "https://steamcommunity.com/market/listings/730/AK-47%20%7C%20Redline%20%28Field-Tested%29",
			expected: &ParsedItem{
				URL:          "https://steamcommunity.com/market/listings/730/AK-47%20%7C%20Redline%20%28Field-Tested%29",
				AppID:        730,
				WearCategory: nil,
			},
		},
		{
			name: "with subdomain",
			url:  "https://xyz.steamcommunity.com/market/listings/570/Immortal%20Tiara",
			expected: &ParsedItem{
				URL:          "https://xyz.steamcommunity.com/market/listings/570/Immortal%20Tiara",
				AppID:        570,
				WearCategory: nil,
			},
		},
		{
			name: "with wear category",
			url:  "https://steamcommunity.com/market/listings/730/G180720B7093004?appid=730&category_730_Exterior=tag_WearCategory3",
			expected: &ParsedItem{
				URL:          "https://steamcommunity.com/market/listings/730/G180720B7093004?appid=730&category_730_Exterior=tag_WearCategory3",
				AppID:        730,
				WearCategory: intPtr(3),
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.ParseSteamItemURL(tt.url)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestParseItemURL_InvalidHost(t *testing.T) {
	client := &Client{}
	invalidURLs := []string{
		"https://example.com/market/listings/730/AK-47",
		"https://steamcommunity.ru/market/listings/730/AK-47",
		"http://notsteam.com/market/listings/730/test",
	}
	for _, url := range invalidURLs {
		t.Run("invalid host: "+url, func(t *testing.T) {
			_, err := client.ParseSteamItemURL(url)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported host")
		})
	}
}

func TestParseItem_invalidPath(t *testing.T) {
	client := &Client{}

	invalidPath := []string{
		"https://steamcommunity.com/",
		"https://steamcommunity.com/market",
		"https://steamcommunity.com/market/listings",
		"https://steamcommunity.com/market/listings/",
		"https://steamcommunity.com/market/listings//item",
		"https://steamcommunity.com/other/listings/730/item",
	}
	for _, url := range invalidPath {
		t.Run("invalid path: "+url, func(t *testing.T) {
			_, err := client.ParseSteamItemURL(url)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid Steam market URL format")
		})
	}
}
func TestParseItemURL_InvalidAppID(t *testing.T) {
	client := &Client{}
	invalidAppIDs := []string{
		"https://steamcommunity.com/market/listings/abc/item",
		"https://steamcommunity.com/market/listings/-1/item",
		"https://steamcommunity.com/market/listings/0/item",
	}

	for _, url := range invalidAppIDs {
		t.Run("invalid appid: "+url, func(t *testing.T) {
			_, err := client.ParseSteamItemURL(url)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid appID")
		})
	}
}

func TestParseItemURL_EmptyOrInvalidHash(t *testing.T) {
	client := &Client{}
	emptyHashUrls := []string{
		"https://steamcommunity.com/market/listings/730/",
		"https://steamcommunity.com/market/listings/730/%00", // invalid encoding
	}

	for _, url := range emptyHashUrls {
		t.Run("empty or invalid hash: "+url, func(t *testing.T) {
			_, err := client.ParseSteamItemURL(url)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "empty market hash name")
		})
	}
}
func intPtr(i int) *int {
	return &i
}
