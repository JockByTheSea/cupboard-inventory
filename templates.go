package main

import (
	"html/template"
	"log"
	"time"
)

var tmpl *template.Template

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

func initTemplates() {
	var err error
	tmpl, err = template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal("Failed to parse template:", err)
	}
}
