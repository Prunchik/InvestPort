package bootstrap

import (
	"investPort/internal/service"
	"investPort/internal/steam"
	"log"
)

func SeedItems(itemService *service.ItemService, client *steam.Client) {
	items := []string{
		"https://steamcommunity.com/market/listings/730/Sealed%20Dead%20Hand%20Terminal",
		"https://steamcommunity.com/market/listings/730/Dreams%20%26%20Nightmares%20Case",
		"https://steamcommunity.com/market/listings/730/Fever%20Case",
		"https://steamcommunity.com/market/listings/730/Danger%20Zone%20Case",
		"https://steamcommunity.com/market/listings/730/Prisma%202%20Case",
		"https://steamcommunity.com/market/listings/730/Snakebite%20Case",
		"https://steamcommunity.com/market/listings/730/Prisma%20Case",
		"https://steamcommunity.com/market/listings/730/Chroma%202%20Case",
	}
	for _, url := range items {
		item, err := client.ParseItemURL(url)
		if err != nil {
			log.Printf("failed to parse url: %v\n", err)
		}
		_, err = itemService.GetOrCreateItem(item)
		if err != nil {
			log.Printf("failed to seed item: %v\n", err)
		}
	}
}
