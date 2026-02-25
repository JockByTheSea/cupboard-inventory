package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

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

const dataFile = "data.json"

var tmpl *template.Template

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

var funcMap = template.FuncMap{
	"isExpired": func(expiry string) bool {
		if expiry == "" {
			return false
		}
		t, err := time.Parse("2006-01-02", expiry)
		if err != nil {
			return false
		}
		return time.Now().After(t)
	},
	"isExpiringSoon": func(expiry string) bool {
		if expiry == "" {
			return false
		}
		t, err := time.Parse("2006-01-02", expiry)
		if err != nil {
			return false
		}
		now := time.Now()
		return !now.After(t) && t.Before(now.Add(7*24*time.Hour))
	},
	"categoryClass": func(category string) string {
		classes := map[string]string{
			"Canned Goods": "cat-canned",
			"Dry Goods":    "cat-dry",
			"Spices":       "cat-spices",
			"Condiments":   "cat-condiments",
			"Baking":       "cat-baking",
			"Snacks":       "cat-snacks",
			"Beverages":    "cat-beverages",
			"Other":        "cat-other",
		}
		if c, ok := classes[category]; ok {
			return c
		}
		return "cat-other"
	},
	"daysInFreezer": func(dateFrozen string) int {
		if dateFrozen == "" {
			return 0
		}
		t, err := time.Parse("2006-01-02", dateFrozen)
		if err != nil {
			return 0
		}
		days := int(time.Since(t).Hours() / 24)
		if days < 0 {
			return 0
		}
		return days
	},
	"freezerAgeClass": func(dateFrozen string) string {
		if dateFrozen == "" {
			return "age-fresh"
		}
		t, err := time.Parse("2006-01-02", dateFrozen)
		if err != nil {
			return "age-fresh"
		}
		days := int(time.Since(t).Hours() / 24)
		switch {
		case days > 90:
			return "age-old"
		case days > 30:
			return "age-medium"
		default:
			return "age-fresh"
		}
	},
}

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

func main() {
	var err error
	tmpl, err = template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal("Failed to parse template:", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/pantry/add", addPantryHandler)
	mux.HandleFunc("/pantry/edit", editPantryHandler)
	mux.HandleFunc("/pantry/delete", deletePantryHandler)
	mux.HandleFunc("/freezer/add", addFreezerHandler)
	mux.HandleFunc("/freezer/edit", editFreezerHandler)
	mux.HandleFunc("/freezer/delete", deleteFreezerHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
