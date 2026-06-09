package models

import "testing"

func TestIsValidSnackType(t *testing.T) {
	if !IsValidSnackType("Candy") {
		t.Fatal("expected Candy to be valid")
	}
	if !IsValidSnackType("Chips/Crackers") {
		t.Fatal("expected Chips/Crackers to be valid")
	}
	if IsValidSnackType("Chocolate") {
		t.Fatal("expected unknown type to be invalid")
	}
}
