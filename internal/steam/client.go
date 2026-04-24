package steam

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
}
type PriceResponse struct {
	Success     bool   `json:"success"`
	LowestPrice string `json:"lowest_price"`
	MedianPrice string `json:"median_price"`
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) GetPrice(marketHashName string, appID int) (float64, error) {
	apiUrl := fmt.Sprintf("https://steamcommunity.com/market/priceoverview/?appid=%d&market_hash_name=%s",
		appID,
		marketHashName)
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return 0, nil
	}
	fmt.Println(marketHashName)
	fmt.Println(url.QueryEscape(marketHashName))
	fmt.Println(apiUrl)

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch steam api: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("steam api returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read body, err: %w", err)
	}

	var marketResponse PriceResponse
	err = json.Unmarshal(body, &marketResponse)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal body, err: %w", err)
	}
	if !marketResponse.Success {
		return 0, fmt.Errorf("steam api returned success = false")
	}
	priceStr := marketResponse.MedianPrice
	if priceStr == "" {
		priceStr = marketResponse.LowestPrice
	}
	if priceStr == "" {
		return 0, fmt.Errorf("no price data available")
	}
	cleaned := strings.TrimSpace(strings.TrimLeft(priceStr, "$€£¥₽"))
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	median, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float, err: %w", err)
	}
	return median, nil
}
