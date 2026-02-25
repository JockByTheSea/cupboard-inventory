package main

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	store, err := loadStore()
	if err != nil {
		http.Error(w, "Failed to load data", http.StatusInternalServerError)
		return
	}

	// Sort pantry items: expiring soonest first, no expiry at the end, then alphabetically
	sort.Slice(store.PantryItems, func(i, j int) bool {
		ei, ej := store.PantryItems[i].Expiry, store.PantryItems[j].Expiry
		if ei == "" && ej == "" {
			return store.PantryItems[i].Name < store.PantryItems[j].Name
		}
		if ei == "" {
			return false
		}
		if ej == "" {
			return true
		}
		return ei < ej
	})

	// Sort freezer meals: oldest first (eat those first), then alphabetically
	sort.Slice(store.FreezerMeals, func(i, j int) bool {
		di, dj := store.FreezerMeals[i].DateFrozen, store.FreezerMeals[j].DateFrozen
		if di == dj {
			return store.FreezerMeals[i].Name < store.FreezerMeals[j].Name
		}
		return di < dj
	})

	if err := tmpl.Execute(w, store); err != nil {
		log.Println("Template error:", err)
	}
}

func addPantryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	store, err := loadStore()
	if err != nil {
		http.Error(w, "Failed to load data", http.StatusInternalServerError)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	store.PantryItems = append(store.PantryItems, PantryItem{
		ID:       store.NextPantryID,
		Name:     name,
		Quantity: strings.TrimSpace(r.FormValue("quantity")),
		Category: r.FormValue("category"),
		Expiry:   r.FormValue("expiry"),
		Notes:    strings.TrimSpace(r.FormValue("notes")),
	})
	store.NextPantryID++
	if err := saveStore(store); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editPantryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	store, err := loadStore()
	if err != nil {
		http.Error(w, "Failed to load data", http.StatusInternalServerError)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	for i := range store.PantryItems {
		if store.PantryItems[i].ID == id {
			store.PantryItems[i].Name = name
			store.PantryItems[i].Quantity = strings.TrimSpace(r.FormValue("quantity"))
			store.PantryItems[i].Category = r.FormValue("category")
			store.PantryItems[i].Expiry = r.FormValue("expiry")
			store.PantryItems[i].Notes = strings.TrimSpace(r.FormValue("notes"))
			break
		}
	}
	if err := saveStore(store); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deletePantryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	store, err := loadStore()
	if err != nil {
		http.Error(w, "Failed to load data", http.StatusInternalServerError)
		return
	}
	items := make([]PantryItem, 0, len(store.PantryItems))
	for _, item := range store.PantryItems {
		if item.ID != id {
			items = append(items, item)
		}
	}
	store.PantryItems = items
	if err := saveStore(store); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func addFreezerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	store, err := loadStore()
	if err != nil {
		http.Error(w, "Failed to load data", http.StatusInternalServerError)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	store.FreezerMeals = append(store.FreezerMeals, FreezerMeal{
		ID:          store.NextMealID,
		Name:        name,
		Portions:    strings.TrimSpace(r.FormValue("portions")),
		DateFrozen:  r.FormValue("date_frozen"),
		Description: strings.TrimSpace(r.FormValue("description")),
	})
	store.NextMealID++
	if err := saveStore(store); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editFreezerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	store, err := loadStore()
	if err != nil {
		http.Error(w, "Failed to load data", http.StatusInternalServerError)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	for i := range store.FreezerMeals {
		if store.FreezerMeals[i].ID == id {
			store.FreezerMeals[i].Name = name
			store.FreezerMeals[i].Portions = strings.TrimSpace(r.FormValue("portions"))
			store.FreezerMeals[i].DateFrozen = r.FormValue("date_frozen")
			store.FreezerMeals[i].Description = strings.TrimSpace(r.FormValue("description"))
			break
		}
	}
	if err := saveStore(store); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteFreezerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	store, err := loadStore()
	if err != nil {
		http.Error(w, "Failed to load data", http.StatusInternalServerError)
		return
	}
	meals := make([]FreezerMeal, 0, len(store.FreezerMeals))
	for _, meal := range store.FreezerMeals {
		if meal.ID != id {
			meals = append(meals, meal)
		}
	}
	store.FreezerMeals = meals
	if err := saveStore(store); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
