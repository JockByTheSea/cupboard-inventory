package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// useTempDB redirects dbFile to a fresh temp file for the duration of a test.
func useTempDB(t *testing.T) {
	t.Helper()
	orig := dbFile
	dbFile = filepath.Join(t.TempDir(), "test.db")
	t.Cleanup(func() { dbFile = orig })
}

func TestInitDB(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		t.Fatalf("initDB: %v", err)
	}

	// Tables must exist (re-running initDB should be idempotent).
	if err := initDB(db); err != nil {
		t.Fatalf("initDB second call: %v", err)
	}
}

func TestLoadStoreEmpty(t *testing.T) {
	useTempDB(t)

	store, err := loadStore()
	if err != nil {
		t.Fatalf("loadStore: %v", err)
	}
	if len(store.PantryItems) != 0 {
		t.Errorf("expected 0 pantry items, got %d", len(store.PantryItems))
	}
	if len(store.FreezerMeals) != 0 {
		t.Errorf("expected 0 freezer meals, got %d", len(store.FreezerMeals))
	}
	if store.NextPantryID != 1 {
		t.Errorf("expected NextPantryID=1, got %d", store.NextPantryID)
	}
	if store.NextMealID != 1 {
		t.Errorf("expected NextMealID=1, got %d", store.NextMealID)
	}
}

func TestSaveAndLoadPantryItems(t *testing.T) {
	useTempDB(t)

	store := &Store{
		PantryItems: []PantryItem{
			{ID: 1, Name: "Rice", Quantity: "2kg", Category: "Dry Goods", Expiry: "2027-01-01", Notes: "bulk"},
			{ID: 2, Name: "Salt", Quantity: "500g", Category: "Spices", Expiry: "", Notes: ""},
		},
		FreezerMeals: []FreezerMeal{},
		NextPantryID: 3,
		NextMealID:   1,
	}

	if err := saveStore(store); err != nil {
		t.Fatalf("saveStore: %v", err)
	}

	loaded, err := loadStore()
	if err != nil {
		t.Fatalf("loadStore: %v", err)
	}
	if len(loaded.PantryItems) != 2 {
		t.Fatalf("expected 2 pantry items, got %d", len(loaded.PantryItems))
	}

	// Build map by ID for order-independent comparison.
	byID := make(map[int]PantryItem)
	for _, item := range loaded.PantryItems {
		byID[item.ID] = item
	}

	rice := byID[1]
	if rice.Name != "Rice" || rice.Quantity != "2kg" || rice.Category != "Dry Goods" ||
		rice.Expiry != "2027-01-01" || rice.Notes != "bulk" {
		t.Errorf("unexpected rice data: %+v", rice)
	}
	salt := byID[2]
	if salt.Name != "Salt" || salt.Quantity != "500g" || salt.Category != "Spices" ||
		salt.Expiry != "" || salt.Notes != "" {
		t.Errorf("unexpected salt data: %+v", salt)
	}
}

func TestSaveAndLoadFreezerMeals(t *testing.T) {
	useTempDB(t)

	store := &Store{
		PantryItems: []PantryItem{},
		FreezerMeals: []FreezerMeal{
			{ID: 1, Name: "Bolognese", Portions: "4", DateFrozen: "2026-01-10", Description: "spicy"},
		},
		NextPantryID: 1,
		NextMealID:   2,
	}

	if err := saveStore(store); err != nil {
		t.Fatalf("saveStore: %v", err)
	}

	loaded, err := loadStore()
	if err != nil {
		t.Fatalf("loadStore: %v", err)
	}
	if len(loaded.FreezerMeals) != 1 {
		t.Fatalf("expected 1 freezer meal, got %d", len(loaded.FreezerMeals))
	}
	m := loaded.FreezerMeals[0]
	if m.ID != 1 || m.Name != "Bolognese" || m.Portions != "4" ||
		m.DateFrozen != "2026-01-10" || m.Description != "spicy" {
		t.Errorf("unexpected meal data: %+v", m)
	}
}

func TestSaveStoreOverwrites(t *testing.T) {
	useTempDB(t)

	// First save two items.
	store := &Store{
		PantryItems: []PantryItem{
			{ID: 1, Name: "Beans", Quantity: "1 can", Category: "Canned Goods"},
			{ID: 2, Name: "Pasta", Quantity: "500g", Category: "Dry Goods"},
		},
		FreezerMeals: []FreezerMeal{},
		NextPantryID: 3,
		NextMealID:   1,
	}
	if err := saveStore(store); err != nil {
		t.Fatalf("saveStore first: %v", err)
	}

	// Second save with only one item – the other must be gone.
	store.PantryItems = []PantryItem{
		{ID: 2, Name: "Pasta", Quantity: "500g", Category: "Dry Goods"},
	}
	if err := saveStore(store); err != nil {
		t.Fatalf("saveStore second: %v", err)
	}

	loaded, err := loadStore()
	if err != nil {
		t.Fatalf("loadStore: %v", err)
	}
	if len(loaded.PantryItems) != 1 {
		t.Fatalf("expected 1 pantry item after overwrite, got %d", len(loaded.PantryItems))
	}
	if loaded.PantryItems[0].Name != "Pasta" {
		t.Errorf("expected Pasta, got %s", loaded.PantryItems[0].Name)
	}
}

func TestLoadStoreNextIDCalculation(t *testing.T) {
	useTempDB(t)

	store := &Store{
		PantryItems: []PantryItem{
			{ID: 5, Name: "A"},
			{ID: 3, Name: "B"},
		},
		FreezerMeals: []FreezerMeal{
			{ID: 7, Name: "X"},
		},
		NextPantryID: 6,
		NextMealID:   8,
	}
	if err := saveStore(store); err != nil {
		t.Fatalf("saveStore: %v", err)
	}

	loaded, err := loadStore()
	if err != nil {
		t.Fatalf("loadStore: %v", err)
	}
	if loaded.NextPantryID != 6 {
		t.Errorf("expected NextPantryID=6, got %d", loaded.NextPantryID)
	}
	if loaded.NextMealID != 8 {
		t.Errorf("expected NextMealID=8, got %d", loaded.NextMealID)
	}
}
