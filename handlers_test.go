package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// setupHandlerTest initialises templates and redirects the DB to a temp file.
func setupHandlerTest(t *testing.T) {
	t.Helper()
	initTemplates()
	useTempDB(t)
}

// ---- index handler ----

func TestIndexHandlerEmpty(t *testing.T) {
	setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	indexHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestIndexHandlerNotFound(t *testing.T) {
	setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	indexHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestIndexHandlerShowsPantryItem(t *testing.T) {
	setupHandlerTest(t)

	// Seed one pantry item.
	store := &Store{
		PantryItems:  []PantryItem{{ID: 1, Name: "Tomatoes", Category: "Canned Goods"}},
		FreezerMeals: []FreezerMeal{},
		NextPantryID: 2,
		NextMealID:   1,
	}
	if err := saveStore(store); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	indexHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Tomatoes") {
		t.Error("response body should contain 'Tomatoes'")
	}
}

// ---- pantry add handler ----

func TestAddPantryHandlerRedirectsOnGet(t *testing.T) {
	setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/pantry/add", nil)
	w := httptest.NewRecorder()
	addPantryHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}
}

func TestAddPantryHandlerEmptyName(t *testing.T) {
	setupHandlerTest(t)

	form := url.Values{"name": {""}}
	req := httptest.NewRequest(http.MethodPost, "/pantry/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	addPantryHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	// Nothing should have been stored.
	store, err := loadStore()
	if err != nil {
		t.Fatal(err)
	}
	if len(store.PantryItems) != 0 {
		t.Errorf("expected 0 items after empty-name add, got %d", len(store.PantryItems))
	}
}

func TestAddPantryHandlerAddsItem(t *testing.T) {
	setupHandlerTest(t)

	form := url.Values{
		"name":     {"Chickpeas"},
		"quantity": {"2 cans"},
		"category": {"Canned Goods"},
		"expiry":   {"2028-06-01"},
		"notes":    {"low sodium"},
	}
	req := httptest.NewRequest(http.MethodPost, "/pantry/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	addPantryHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	store, err := loadStore()
	if err != nil {
		t.Fatal(err)
	}
	if len(store.PantryItems) != 1 {
		t.Fatalf("expected 1 pantry item, got %d", len(store.PantryItems))
	}
	item := store.PantryItems[0]
	if item.Name != "Chickpeas" || item.Quantity != "2 cans" || item.Category != "Canned Goods" ||
		item.Expiry != "2028-06-01" || item.Notes != "low sodium" {
		t.Errorf("unexpected item: %+v", item)
	}
}

// ---- pantry edit handler ----

func TestEditPantryHandlerRedirectsOnGet(t *testing.T) {
	setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/pantry/edit", nil)
	w := httptest.NewRecorder()
	editPantryHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}
}

func TestEditPantryHandlerUpdatesItem(t *testing.T) {
	setupHandlerTest(t)

	// Seed an item to edit.
	if err := saveStore(&Store{
		PantryItems:  []PantryItem{{ID: 1, Name: "Old Name", Quantity: "1"}},
		FreezerMeals: []FreezerMeal{},
		NextPantryID: 2,
		NextMealID:   1,
	}); err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"id":       {"1"},
		"name":     {"New Name"},
		"quantity": {"5"},
		"category": {"Snacks"},
		"expiry":   {"2029-01-01"},
		"notes":    {"updated"},
	}
	req := httptest.NewRequest(http.MethodPost, "/pantry/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	editPantryHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	store, err := loadStore()
	if err != nil {
		t.Fatal(err)
	}
	if len(store.PantryItems) != 1 {
		t.Fatalf("expected 1 item, got %d", len(store.PantryItems))
	}
	item := store.PantryItems[0]
	if item.Name != "New Name" || item.Quantity != "5" || item.Category != "Snacks" ||
		item.Expiry != "2029-01-01" || item.Notes != "updated" {
		t.Errorf("unexpected item after edit: %+v", item)
	}
}

func TestEditPantryHandlerEmptyName(t *testing.T) {
	setupHandlerTest(t)

	if err := saveStore(&Store{
		PantryItems:  []PantryItem{{ID: 1, Name: "Keep Me"}},
		FreezerMeals: []FreezerMeal{},
		NextPantryID: 2,
		NextMealID:   1,
	}); err != nil {
		t.Fatal(err)
	}

	form := url.Values{"id": {"1"}, "name": {""}}
	req := httptest.NewRequest(http.MethodPost, "/pantry/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	editPantryHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	// Name must be unchanged.
	store, _ := loadStore()
	if store.PantryItems[0].Name != "Keep Me" {
		t.Errorf("name should not change when empty name submitted")
	}
}

// ---- pantry delete handler ----

func TestDeletePantryHandlerRedirectsOnGet(t *testing.T) {
	setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/pantry/delete", nil)
	w := httptest.NewRecorder()
	deletePantryHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}
}

func TestDeletePantryHandlerDeletesItem(t *testing.T) {
	setupHandlerTest(t)

	if err := saveStore(&Store{
		PantryItems:  []PantryItem{{ID: 1, Name: "Delete Me"}, {ID: 2, Name: "Keep Me"}},
		FreezerMeals: []FreezerMeal{},
		NextPantryID: 3,
		NextMealID:   1,
	}); err != nil {
		t.Fatal(err)
	}

	form := url.Values{"id": {"1"}}
	req := httptest.NewRequest(http.MethodPost, "/pantry/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	deletePantryHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	store, err := loadStore()
	if err != nil {
		t.Fatal(err)
	}
	if len(store.PantryItems) != 1 {
		t.Fatalf("expected 1 item after delete, got %d", len(store.PantryItems))
	}
	if store.PantryItems[0].Name != "Keep Me" {
		t.Errorf("wrong item remaining: %s", store.PantryItems[0].Name)
	}
}

// ---- freezer add handler ----

func TestAddFreezerHandlerRedirectsOnGet(t *testing.T) {
	setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/freezer/add", nil)
	w := httptest.NewRecorder()
	addFreezerHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}
}

func TestAddFreezerHandlerAddsMeal(t *testing.T) {
	setupHandlerTest(t)

	form := url.Values{
		"name":        {"Lasagne"},
		"portions":    {"6"},
		"date_frozen": {"2026-01-15"},
		"description": {"beef and béchamel"},
	}
	req := httptest.NewRequest(http.MethodPost, "/freezer/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	addFreezerHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	store, err := loadStore()
	if err != nil {
		t.Fatal(err)
	}
	if len(store.FreezerMeals) != 1 {
		t.Fatalf("expected 1 meal, got %d", len(store.FreezerMeals))
	}
	m := store.FreezerMeals[0]
	if m.Name != "Lasagne" || m.Portions != "6" || m.DateFrozen != "2026-01-15" || m.Description != "beef and béchamel" {
		t.Errorf("unexpected meal: %+v", m)
	}
}

func TestAddFreezerHandlerEmptyName(t *testing.T) {
	setupHandlerTest(t)

	form := url.Values{"name": {""}}
	req := httptest.NewRequest(http.MethodPost, "/freezer/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	addFreezerHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	store, _ := loadStore()
	if len(store.FreezerMeals) != 0 {
		t.Errorf("expected 0 meals after empty-name add, got %d", len(store.FreezerMeals))
	}
}

// ---- freezer edit handler ----

func TestEditFreezerHandlerRedirectsOnGet(t *testing.T) {
	setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/freezer/edit", nil)
	w := httptest.NewRecorder()
	editFreezerHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}
}

func TestEditFreezerHandlerUpdatesMeal(t *testing.T) {
	setupHandlerTest(t)

	if err := saveStore(&Store{
		PantryItems:  []PantryItem{},
		FreezerMeals: []FreezerMeal{{ID: 1, Name: "Old Stew", Portions: "2", DateFrozen: "2026-01-01"}},
		NextPantryID: 1,
		NextMealID:   2,
	}); err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"id":          {"1"},
		"name":        {"New Stew"},
		"portions":    {"4"},
		"date_frozen": {"2026-02-01"},
		"description": {"hearty"},
	}
	req := httptest.NewRequest(http.MethodPost, "/freezer/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	editFreezerHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	store, err := loadStore()
	if err != nil {
		t.Fatal(err)
	}
	m := store.FreezerMeals[0]
	if m.Name != "New Stew" || m.Portions != "4" || m.DateFrozen != "2026-02-01" || m.Description != "hearty" {
		t.Errorf("unexpected meal after edit: %+v", m)
	}
}

// ---- freezer delete handler ----

func TestDeleteFreezerHandlerRedirectsOnGet(t *testing.T) {
	setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/freezer/delete", nil)
	w := httptest.NewRecorder()
	deleteFreezerHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}
}

func TestDeleteFreezerHandlerDeletesMeal(t *testing.T) {
	setupHandlerTest(t)

	if err := saveStore(&Store{
		PantryItems:  []PantryItem{},
		FreezerMeals: []FreezerMeal{{ID: 1, Name: "Gone"}, {ID: 2, Name: "Stays"}},
		NextPantryID: 1,
		NextMealID:   3,
	}); err != nil {
		t.Fatal(err)
	}

	form := url.Values{"id": {"1"}}
	req := httptest.NewRequest(http.MethodPost, "/freezer/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	deleteFreezerHandler(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}

	store, err := loadStore()
	if err != nil {
		t.Fatal(err)
	}
	if len(store.FreezerMeals) != 1 {
		t.Fatalf("expected 1 meal after delete, got %d", len(store.FreezerMeals))
	}
	if store.FreezerMeals[0].Name != "Stays" {
		t.Errorf("wrong meal remaining: %s", store.FreezerMeals[0].Name)
	}
}
