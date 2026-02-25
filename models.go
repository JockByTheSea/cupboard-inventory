package main

// PantryItem represents an item stored in the pantry.
type PantryItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
	Category string `json:"category"`
	Expiry   string `json:"expiry"`
	Notes    string `json:"notes"`
}

// FreezerMeal represents a leftover meal stored in the freezer.
type FreezerMeal struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Portions    string `json:"portions"`
	DateFrozen  string `json:"date_frozen"`
	Description string `json:"description"`
}

// Store holds all application data.
type Store struct {
	PantryItems  []PantryItem  `json:"pantry_items"`
	FreezerMeals []FreezerMeal `json:"freezer_meals"`
	NextPantryID int           `json:"next_pantry_id"`
	NextMealID   int           `json:"next_meal_id"`
}
