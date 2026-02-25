package main

import (
	"testing"
	"time"
)

// Retrieve the typed functions from funcMap for direct testing.

var (
	fnIsExpired      = funcMap["isExpired"].(func(string) bool)
	fnIsExpiringSoon = funcMap["isExpiringSoon"].(func(string) bool)
	fnCategoryClass  = funcMap["categoryClass"].(func(string) string)
	fnDaysInFreezer  = funcMap["daysInFreezer"].(func(string) int)
	fnFreezerAge     = funcMap["freezerAgeClass"].(func(string) string)
)

// ---- isExpired ----

func TestIsExpiredEmpty(t *testing.T) {
	if fnIsExpired("") {
		t.Error("empty expiry should not be expired")
	}
}

func TestIsExpiredInvalidDate(t *testing.T) {
	if fnIsExpired("not-a-date") {
		t.Error("invalid date should not be expired")
	}
}

func TestIsExpiredPastDate(t *testing.T) {
	if !fnIsExpired("2000-01-01") {
		t.Error("past date should be expired")
	}
}

func TestIsExpiredFutureDate(t *testing.T) {
	if fnIsExpired("2999-12-31") {
		t.Error("future date should not be expired")
	}
}

// ---- isExpiringSoon ----

func TestIsExpiringSoonEmpty(t *testing.T) {
	if fnIsExpiringSoon("") {
		t.Error("empty expiry should not be expiring soon")
	}
}

func TestIsExpiringSoonInvalidDate(t *testing.T) {
	if fnIsExpiringSoon("bad") {
		t.Error("invalid date should not be expiring soon")
	}
}

func TestIsExpiringSoonAlreadyExpired(t *testing.T) {
	if fnIsExpiringSoon("2000-01-01") {
		t.Error("already-expired item should not be 'expiring soon'")
	}
}

func TestIsExpiringSoonWithinWeek(t *testing.T) {
	soon := time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02")
	if !fnIsExpiringSoon(soon) {
		t.Errorf("date %s (3 days away) should be expiring soon", soon)
	}
}

func TestIsExpiringSoonFarFuture(t *testing.T) {
	if fnIsExpiringSoon("2999-12-31") {
		t.Error("far-future date should not be expiring soon")
	}
}

// ---- categoryClass ----

func TestCategoryClassKnown(t *testing.T) {
	cases := map[string]string{
		"Canned Goods": "cat-canned",
		"Dry Goods":    "cat-dry",
		"Spices":       "cat-spices",
		"Condiments":   "cat-condiments",
		"Baking":       "cat-baking",
		"Snacks":       "cat-snacks",
		"Beverages":    "cat-beverages",
		"Other":        "cat-other",
	}
	for cat, want := range cases {
		if got := fnCategoryClass(cat); got != want {
			t.Errorf("categoryClass(%q) = %q, want %q", cat, got, want)
		}
	}
}

func TestCategoryClassUnknown(t *testing.T) {
	if got := fnCategoryClass("Mystery"); got != "cat-other" {
		t.Errorf("unknown category should map to cat-other, got %q", got)
	}
}

func TestCategoryClassEmpty(t *testing.T) {
	if got := fnCategoryClass(""); got != "cat-other" {
		t.Errorf("empty category should map to cat-other, got %q", got)
	}
}

// ---- daysInFreezer ----

func TestDaysInFreezerEmpty(t *testing.T) {
	if d := fnDaysInFreezer(""); d != 0 {
		t.Errorf("empty date should return 0, got %d", d)
	}
}

func TestDaysInFreezerInvalidDate(t *testing.T) {
	if d := fnDaysInFreezer("not-a-date"); d != 0 {
		t.Errorf("invalid date should return 0, got %d", d)
	}
}

func TestDaysInFreezerToday(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	d := fnDaysInFreezer(today)
	if d < 0 || d > 1 {
		t.Errorf("today's date should return 0 or 1 days, got %d", d)
	}
}

func TestDaysInFreezerPast(t *testing.T) {
	past := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	d := fnDaysInFreezer(past)
	if d < 9 || d > 11 {
		t.Errorf("10-day-old date should return ~10 days, got %d", d)
	}
}

func TestDaysInFreezerFuture(t *testing.T) {
	future := time.Now().AddDate(0, 0, 5).Format("2006-01-02")
	if d := fnDaysInFreezer(future); d != 0 {
		t.Errorf("future date should return 0, got %d", d)
	}
}

// ---- freezerAgeClass ----

func TestFreezerAgeClassEmpty(t *testing.T) {
	if got := fnFreezerAge(""); got != "age-fresh" {
		t.Errorf("empty date should return age-fresh, got %q", got)
	}
}

func TestFreezerAgeClassFresh(t *testing.T) {
	recent := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	if got := fnFreezerAge(recent); got != "age-fresh" {
		t.Errorf("10-day-old date should return age-fresh, got %q", got)
	}
}

func TestFreezerAgeClassMedium(t *testing.T) {
	medium := time.Now().AddDate(0, 0, -45).Format("2006-01-02")
	if got := fnFreezerAge(medium); got != "age-medium" {
		t.Errorf("45-day-old date should return age-medium, got %q", got)
	}
}

func TestFreezerAgeClassOld(t *testing.T) {
	old := time.Now().AddDate(0, 0, -100).Format("2006-01-02")
	if got := fnFreezerAge(old); got != "age-old" {
		t.Errorf("100-day-old date should return age-old, got %q", got)
	}
}
