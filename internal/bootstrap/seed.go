package bootstrap

import (
	"investPort/internal/service"
	"log"
)

func SeedItems(itemService *service.ItemService) {
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
		_, err := itemService.GetOrCreateByURL(url)
		if err != nil {
			log.Println("failed to seed item: ", err)
		}
	}
}
