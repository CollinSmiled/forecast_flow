package main

import "testing"

func TestRequiredPositiveInt64(t *testing.T) {
	t.Setenv("LOCATION_ID", "42")

	value, err := requiredPositiveInt64("LOCATION_ID")
	if err != nil {
		t.Fatalf("parse location ID: %v", err)
	}

	if value != 42 {
		t.Errorf("location ID = %d, want 42", value)
	}
}

func TestRequiredPositiveInt64RejectsInvalidValue(t *testing.T) {
	t.Setenv("LOCATION_ID", "not-a-number")

	if _, err := requiredPositiveInt64("LOCATION_ID"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestPositiveIntOrDefault(t *testing.T) {
	t.Setenv("FORECAST_DAYS", "")

	value, err := positiveIntOrDefault("FORECAST_DAYS", 10)
	if err != nil {
		t.Fatalf("parse forecast days: %v", err)
	}

	if value != 10 {
		t.Errorf("forecast days = %d, want 10", value)
	}
}
