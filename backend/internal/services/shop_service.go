package services

import (
	"fmt"
	"tictactoe/internal/store"
	"time"
)

type ShopItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        int    `json:"cost"`
	Type        string `json:"type"` // "skin"
}

type ShopService struct {
	Store *store.UserStore
}

var Catalog = []ShopItem{
	{ID: "skin_neon", Name: "Neon", Description: "Cyberpunk vibes", Cost: 100, Type: "skin"},
	{ID: "skin_retro", Name: "Retro", Description: "8-bit classic", Cost: 100, Type: "skin"},
	{ID: "skin_gold", Name: "Gold", Description: "Luxury finish", Cost: 250, Type: "skin"},
}

func NewShopService(store *store.UserStore) *ShopService {
	return &ShopService{Store: store}
}

func (s *ShopService) GetCatalog() []ShopItem {
	return Catalog
}

func (s *ShopService) GetShopInfo(nickname string) (map[string]interface{}, error) {
	user, _, err := s.Store.GetUserByNickname(nickname)
	if err != nil {
		return nil, err
	}

	coins, activeSkin, err := s.Store.GetUserShopData(user.ID)
	if err != nil {
		return nil, err
	}

	inventory, err := s.Store.GetInventory(user.ID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"coins":       coins,
		"active_skin": activeSkin,
		"inventory":   inventory,
		"catalog":     Catalog,
	}, nil
}

func (s *ShopService) BuyItem(nickname string, itemID string) error {
	user, _, err := s.Store.GetUserByNickname(nickname)
	if err != nil {
		return err
	}

	// Find item in catalog
	var item *ShopItem
	for _, i := range Catalog {
		if i.ID == itemID {
			item = &i
			break
		}
	}
	if item == nil {
		return fmt.Errorf("item not found")
	}

	return s.Store.PurchaseItem(user.ID, itemID, item.Cost)
}

func (s *ShopService) WatchAd(nickname string) error {
	user, _, err := s.Store.GetUserByNickname(nickname)
	if err != nil {
		return err
	}

	// Mock delay
	time.Sleep(5 * time.Second)
	return s.Store.AddCoins(user.ID, 50)
}

func (s *ShopService) EquipItem(nickname string, itemID string) error {
	user, _, err := s.Store.GetUserByNickname(nickname)
	if err != nil {
		return err
	}

	// Verify item exists in catalog (or is default)
	if itemID != "default" {
		found := false
		for _, i := range Catalog {
			if i.ID == itemID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid item")
		}
	}
	return s.Store.EquipSkin(user.ID, itemID)
}
