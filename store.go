package main

import (
	"encoding/json"
	"os"
)

const dataFile = "data.json"

func loadStore() (*Store, error) {
	store := &Store{
		PantryItems:  []PantryItem{},
		FreezerMeals: []FreezerMeal{},
		NextPantryID: 1,
		NextMealID:   1,
	}
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, store); err != nil {
		return nil, err
	}
	if store.PantryItems == nil {
		store.PantryItems = []PantryItem{}
	}
	if store.FreezerMeals == nil {
		store.FreezerMeals = []FreezerMeal{}
	}
	return store, nil
}

func saveStore(store *Store) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, data, 0644)
}
