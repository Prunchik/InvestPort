package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"golang.org/x/time/rate"
)

type Client struct {
	httpClient *http.Client
	limiter    *rate.Limiter
}
type PriceResponse struct {
	Success     bool   `json:"success"`
	LowestPrice string `json:"lowest_price"`
	MedianPrice string `json:"median_price"`
}
type ParsedItem struct {
	AppId    int
	HashName string
	Url      string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    100,
				MaxConnsPerHost: 10,
				IdleConnTimeout: 30 * time.Second,
			},
		},

		limiter: rate.NewLimiter(rate.Every(4*time.Second), 1),
	}
}

func (c *Client) GetPrice(ctx context.Context, marketHashName string, appID int) (float64, error) {

	jitter := time.Duration(rand.Intn(2000)) * time.Millisecond

	select {
	case <-time.After(jitter):
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return 0, err
	}

	itemName, err := url.QueryUnescape(marketHashName) //TODO надо сделать норм парсер
	if err != nil {
		return 0, err
	}

	params := url.Values{}
	params.Set("appid", strconv.Itoa(appID))
	params.Set("market_hash_name", itemName)

	apiURL := &url.URL{
		Scheme:   "https",
		Host:     "steamcommunity.com",
		Path:     "/market/priceoverview",
		RawQuery: params.Encode(),
	}

	var price float64

	operation := func() error {

		price, err = c.doRequest(ctx, apiURL.String(), itemName)
		return err
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second
	b.MaxInterval = 7 * time.Second

	err = backoff.RetryNotify(
		operation,
		backoff.WithContext(b, ctx),
		func(err error, duration time.Duration) {
			fmt.Printf("Attempt failed: %v, retry in %v", err, duration)
		})

	if err != nil {
		return 0, fmt.Errorf("max retries exceeded: %w", err)
	}

	fmt.Printf("Price fetched: %.2f for %s\n", price, marketHashName)

	return price, nil

}

func (c *Client) doRequest(ctx context.Context, apiURL string, itemName string) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return 0, backoff.Permanent(err)
	}

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0",
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return 0, fmt.Errorf("rate limited")
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return 0, fmt.Errorf("steam server error: %d", resp.StatusCode)
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return 0, backoff.Permanent(
			fmt.Errorf("steam api returned status: %d", resp.StatusCode),
		)
	}

	var marketResponse PriceResponse

	err = json.NewDecoder(resp.Body).Decode(&marketResponse)
	if err != nil {
		return 0, backoff.Permanent(err)
	}

	if !marketResponse.Success {
		return 0, backoff.Permanent(fmt.Errorf("steam success = false"))
	}

	priceStr := marketResponse.MedianPrice
	if priceStr == "" {
		priceStr = marketResponse.LowestPrice
	}

	if priceStr == "" {
		return 0, backoff.Permanent(fmt.Errorf("no price data for item: %v", itemName))
	}

	priceStr = cleanPrice(priceStr)

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, backoff.Permanent(err)
	}
	return price, nil

}

func cleanPrice(price string) string {
	price = strings.TrimSpace(price)

	replacer := strings.NewReplacer(
		"$", "",
		"€", "",
		"£", "",
		"¥", "",
		"₽", "",
		"руб.", "",
	)

	price = replacer.Replace(price)

	price = strings.TrimSpace(price)

	if strings.Count(price, ",") == 1 &&
		!strings.Contains(price, ".") {
		price = strings.Replace(price, ",", ".", 1)
	}

	if strings.Count(price, ",") > 0 &&
		strings.Contains(price, ".") {
		price = strings.ReplaceAll(price, ",", "")
	}

	return strings.TrimSpace(price)
}
