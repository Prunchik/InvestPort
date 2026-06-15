package steam

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	steamSuffix       = " - Steam Community Market"
	title             = "title"
	cs2Exterior       = "category_730_Exterior"
	cs2Tag            = "tag_WearCategory"
	cs2AppID          = 730
	stickerSlabPrefix = "Sticker Slab | "
)

func (c *Client) ResolveSteamMarketItem(urlString string) (*ParsedItem, error) {

	itemName, err := c.ResolveItemName(context.Background(), urlString) //TODO прокинуть ctx
	if err != nil {
		return nil, err
	}

	item, err := c.ParseSteamItemURL(urlString)
	if err != nil {
		return nil, err
	}

	if item.WearCategory != nil {
		if name, ok := WearCategoryName[*item.WearCategory]; ok {
			itemName = itemName + " (" + name + ")"
		}
	}

	item.Name = itemName
	item.HashName = itemName

	return item, nil
}

func (c *Client) ParseSteamItemURL(urlString string) (*ParsedItem, error) {

	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if !strings.HasSuffix(parsedUrl.Host, "steamcommunity.com") {
		return nil, fmt.Errorf("unsupported host: %s", parsedUrl.Host)
	}

	parts := strings.Split(parsedUrl.Path, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid URL path: too short (%s)", parsedUrl.Path)
	}

	if parts[1] != "market" || parts[2] != "listings" {
		return nil, fmt.Errorf("invalid Steam market URL format: expected /market/listings/{app_id}/")
	}

	appIdStr := parts[3]
	if appIdStr == "" {
		return nil, fmt.Errorf("invalid App ID in the link: empty")
	}
	appID, err := strconv.Atoi(appIdStr)

	if err != nil {
		return nil, fmt.Errorf("invalid appId format")
	}

	var wearCategoryTag string
	var wearCategory *int

	if appID == cs2AppID {

		query, err := url.ParseQuery(parsedUrl.RawQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to read query params: %v", err)
		}
		wearCategoryTag = query.Get(cs2Exterior)
		fmt.Println(wearCategoryTag, query, appID)

		if wearCategoryTag != "" {

			wearCategoryStr := strings.TrimPrefix(wearCategoryTag, cs2Tag)

			switch wearCategoryStr {
			case "0", "1", "2", "3", "4":
				num, err := strconv.Atoi(wearCategoryStr)
				wearCategory = &num
				if err != nil {
					return nil, fmt.Errorf("invalid wear category format: %v", wearCategoryStr)
				}
			default:
				return nil, fmt.Errorf("wrong wear category")
			}
		}
	}
	item := &ParsedItem{
		WearCategory: wearCategory,
		URL:          urlString,
		AppID:        appID,
	}
	return item, nil
}

// ResolveItemName TODO можно добавить таблицу candidate items добавлять туда и проверять на существование
func (c *Client) ResolveItemName(ctx context.Context, urlString string) (string, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlString, nil)

	if err != nil {
		return "", fmt.Errorf("failed to create a request, err: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible)")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch url, err: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("steam api returned status %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	itemTitle := doc.Find(title).Text()

	itemName := strings.TrimSuffix(itemTitle, steamSuffix)
	itemName = strings.TrimPrefix(itemName, stickerSlabPrefix) // TODO не игнорировать sticker slab

	itemName = strings.TrimSpace(itemName)

	if itemName == "" {
		return "", fmt.Errorf("item name is not found")
	}

	if itemName == "Market Item" || itemName == "Steam Community Market" {
		return "", fmt.Errorf("failed to resolve item name from HTML title: %s", itemName)
	}
	return itemName, nil
}
