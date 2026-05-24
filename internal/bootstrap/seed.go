package bootstrap

import (
	"investPort/internal/service"
	"investPort/internal/steam"
	"log"
)

func SeedItems(itemService *service.ItemService, client *steam.Client) {
	items := []string{
		"https://steamcommunity.com/market/listings/730/Sealed%20Dead%20Hand%20Terminal",
		//"https://steamcommunity.com/market/listings/730/G18AA203004",
		"https://steamcommunity.com/market/listings/730/G180720C5093004?appid=730&category_730_Exterior=tag_WearCategory3&category_730_Quality=tag_strange",
	}
	for _, url := range items {
		item, err := client.ResolveSteamMarketItem(url)
		if err != nil {
			log.Printf("failed to parse url: %v\n", err)
		}
		_, err = itemService.GetOrCreateItem(item)
		if err != nil {
			log.Printf("failed to seed item: %v\n", err)
		}
	}
}
