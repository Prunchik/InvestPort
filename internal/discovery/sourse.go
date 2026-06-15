package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type DiscoveredItem struct {
	MarketHashName string `json:"market_hash_name"`
	Name           string `json:"name"`
	ImageURL       string `json:"image"`
	WearCategory   *int   `json:"wear_category"`
}

type SteamSource struct{}

type Source interface {
	fetch(ctx context.Context) ([]DiscoveredItem, error)
}

func NewSteamSource() *SteamSource {
	return &SteamSource{}
}

func (s *SteamSource) fetch(ctx context.Context) ([]DiscoveredItem, error) {

	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		"https://raw.githubusercontent.com/ByMykel/CSGO-API/main/public/api/en/skins_not_grouped.json",
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch data: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read response body")
	}
	if len(body) == 0 {
		return nil, errors.New("empty response body")
	}

	var items []DiscoveredItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, errors.New("failed to unmarshal JSON: " + err.Error())
	}

	return items, nil

}
