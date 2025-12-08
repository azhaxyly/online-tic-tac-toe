package store

import (
	"fmt"
)

// AddCoins adds (or subtracts if negative) coins to a user's balance.
func (s *UserStore) AddCoins(userID int, amount int) error {
	query := `UPDATE users SET coins = coins + $1 WHERE id = $2`
	_, err := s.DB.Exec(query, amount, userID)
	if err != nil {
		return fmt.Errorf("add coins: %w", err)
	}
	return nil
}

// PurchaseItem handles the transaction of buying an item:
// 1. Checks balance and item ownership (via inventory constraint).
// 2. Deducts coins.
// 3. Adds item to inventory.
func (s *UserStore) PurchaseItem(userID int, itemID string, cost int) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Check balance and lock row
	var currentCoins int
	err = tx.QueryRow(`SELECT coins FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&currentCoins)
	if err != nil {
		return fmt.Errorf("get balance: %w", err)
	}

	if currentCoins < cost {
		return fmt.Errorf("insufficient funds")
	}

	// 2. Deduct coins
	_, err = tx.Exec(`UPDATE users SET coins = coins - $1 WHERE id = $2`, cost, userID)
	if err != nil {
		return fmt.Errorf("deduct coins: %w", err)
	}

	// 3. Add to inventory
	_, err = tx.Exec(`INSERT INTO inventory (user_id, item_id) VALUES ($1, $2)`, userID, itemID)
	if err != nil {
		// Likely duplicate key violation if already owns item
		return fmt.Errorf("add to inventory: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetInventory returns a list of item IDs owned by the user.
func (s *UserStore) GetInventory(userID int) ([]string, error) {
	rows, err := s.DB.Query(`SELECT item_id FROM inventory WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("query inventory: %w", err)
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var itemID string
		if err := rows.Scan(&itemID); err != nil {
			return nil, err
		}
		items = append(items, itemID)
	}
	return items, nil
}

// EquipSkin updates the user's active skin.
// It verifies that the user owns the skin (unless it's "default").
func (s *UserStore) EquipSkin(userID int, skinID string) error {
	// If it's not default, check ownership
	if skinID != "default" {
		var exists bool
		err := s.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM inventory WHERE user_id = $1 AND item_id = $2)`, userID, skinID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check ownership: %w", err)
		}
		if !exists {
			return fmt.Errorf("user does not own skin: %s", skinID)
		}
	}

	_, err := s.DB.Exec(`UPDATE users SET active_skin = $1 WHERE id = $2`, skinID, userID)
	if err != nil {
		return fmt.Errorf("update active skin: %w", err)
	}
	return nil
}

// GetUserShopData returns coins and active skin for shop display
func (s *UserStore) GetUserShopData(userID int) (int, string, error) {
	var coins int
	var activeSkin string
	err := s.DB.QueryRow(`SELECT coins, active_skin FROM users WHERE id = $1`, userID).Scan(&coins, &activeSkin)
	if err != nil {
		return 0, "", fmt.Errorf("get user shop data: %w", err)
	}
	return coins, activeSkin, nil
}
