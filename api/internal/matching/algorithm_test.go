package matching

import (
	"math/rand"
	"testing"

	"github.com/google/uuid"
)

func TestPairCrossCountry(t *testing.T) {
	us := uuid.New()
	ca := uuid.New()
	gb := uuid.New()
	fr := uuid.New()

	users := []eligibleUser{
		{id: us, country: "US"},
		{id: ca, country: "CA"},
		{id: gb, country: "GB"},
		{id: fr, country: "FR"},
	}

	pairs := pairCrossCountry(users, rand.New(rand.NewSource(1)))
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	for _, pair := range pairs {
		countries := map[string]bool{}
		for _, u := range users {
			if u.id == pair[0] || u.id == pair[1] {
				countries[u.country] = true
			}
		}
		if len(countries) != 2 {
			t.Fatalf("pair %v does not span two countries", pair)
		}
	}
}

func TestPairCrossCountrySkipsSameCountryOnlyPool(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()

	users := []eligibleUser{
		{id: a, country: "US"},
		{id: b, country: "US"},
		{id: c, country: "CA"},
	}

	pairs := pairCrossCountry(users, rand.New(rand.NewSource(2)))
	if len(pairs) != 1 {
		t.Fatalf("expected 1 cross-country pair, got %d", len(pairs))
	}
}

func TestSameCountry(t *testing.T) {
	if !sameCountry("US", "us") {
		t.Fatal("expected US and us to match")
	}
	if sameCountry("US", "CA") {
		t.Fatal("expected US and CA to differ")
	}
}
