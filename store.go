package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

const dbFile = "data.db"

func openDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}
	if err := initDB(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS pantry_items (
			id       INTEGER PRIMARY KEY,
			name     TEXT NOT NULL,
			quantity TEXT,
			category TEXT,
			expiry   TEXT,
			notes    TEXT
		);
		CREATE TABLE IF NOT EXISTS freezer_meals (
			id          INTEGER PRIMARY KEY,
			name        TEXT NOT NULL,
			portions    TEXT,
			date_frozen TEXT,
			description TEXT
		);
	`)
	return err
}

func loadStore() (*Store, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	store := &Store{
		PantryItems:  []PantryItem{},
		FreezerMeals: []FreezerMeal{},
		NextPantryID: 1,
		NextMealID:   1,
	}

	rows, err := db.Query("SELECT id, name, quantity, category, expiry, notes FROM pantry_items")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item PantryItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &item.Category, &item.Expiry, &item.Notes); err != nil {
			return nil, err
		}
		store.PantryItems = append(store.PantryItems, item)
		if item.ID >= store.NextPantryID {
			store.NextPantryID = item.ID + 1
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows2, err := db.Query("SELECT id, name, portions, date_frozen, description FROM freezer_meals")
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var meal FreezerMeal
		if err := rows2.Scan(&meal.ID, &meal.Name, &meal.Portions, &meal.DateFrozen, &meal.Description); err != nil {
			return nil, err
		}
		store.FreezerMeals = append(store.FreezerMeals, meal)
		if meal.ID >= store.NextMealID {
			store.NextMealID = meal.ID + 1
		}
	}
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	return store, nil
}

func saveStore(store *Store) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM pantry_items"); err != nil {
		return err
	}
	for _, item := range store.PantryItems {
		if _, err := tx.Exec(
			"INSERT INTO pantry_items (id, name, quantity, category, expiry, notes) VALUES (?, ?, ?, ?, ?, ?)",
			item.ID, item.Name, item.Quantity, item.Category, item.Expiry, item.Notes,
		); err != nil {
			return err
		}
	}

	if _, err := tx.Exec("DELETE FROM freezer_meals"); err != nil {
		return err
	}
	for _, meal := range store.FreezerMeals {
		if _, err := tx.Exec(
			"INSERT INTO freezer_meals (id, name, portions, date_frozen, description) VALUES (?, ?, ?, ?, ?)",
			meal.ID, meal.Name, meal.Portions, meal.DateFrozen, meal.Description,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
