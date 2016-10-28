package main

import (
	"fmt"
	"testing"
)

func TestLoadCards(t *testing.T) {
	cards, err := loadCards("../AllCards.json")

	if err != nil {
		t.Fatalf("Loading cards failed: %s", err.Error())
	}

	fmt.Printf("%v\n", cards["Bloodsoaked Champion"])
}

func TestLoadSets(t *testing.T) {
	sets, err := loadSets("../AllSets.json")

	if err != nil {
		t.Fatalf("Loading sets failed: %s", err.Error())
	}

	fmt.Printf("%v\n", sets["LEA"].Cards[0])
}

