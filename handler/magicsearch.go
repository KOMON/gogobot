package handler

import (
	"fmt"
	"strings"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)


type MtgSearchResponder struct {
	db *sql.DB
	matches []string
}

type Query struct {
	Name string
	ManaCost string
	CMC string
	Colors []string
	ColorID []string
	Supertypes []string
	Types []string
	Subtypes []string
	Rarity string
	Text string
	Flavor string
	Artist string
	Power string
	Toughness string
	Loyalty string
	Set string
	Block string
}

func NewMtgSearch() MtgSearchResponder {
	db, err := sql.Open("sqlite3", "mtg.db")
	if err != nil {
		log.Fatal(err)
	}
	
	return MtgSearchResponder{
		db: db,
		matches: []string{},
	}
}

func (msr *MtgSearchResponder) Match(s string) bool {
	msr.matches = []string{}

	for i:= strings.Index(s, "[["); i != -1; i = strings.Index(s, "[[") {
		fmt.Println("Looping", i, s)
		j := strings.Index(s, "]]")

		if i != 0 && s[i-1] == '#' {
			s = s[j+2:]
			continue
		}
		msr.matches = append(msr.matches, s[i+2:j])
		s = s[j+2:]
	}

	if len(msr.matches) > 0 {
		return true
	}

	return false
}

func (msr MtgSearchResponder) Respond() (string, error) {
	findCard := `select name, mana_cost, card_text, multiverse_id from cards where id = (select id from virt_cards where name match ? and multiverse_id != 0) order by release_date limit 1`
	findCardBySet := `select name, mana_cost, card_text, multiverse_id from cards inner join set_card where cards.id = (select id from virt_cards where name match ? and multiverse_id != 0) and set_code = ? limit 1`
	tx, err := msr.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	findStmt, err := tx.Prepare(findCard)
	if err != nil {
		log.Fatal(err)
	}
	defer findStmt.Close()

	findBySetStmt, err := tx.Prepare(findCardBySet)
	if err != nil {
		log.Fatal(err)
	}
	defer findBySetStmt.Close()

	response := ""
	multi := len(msr.matches) > 1
	for _, match := range msr.matches {
		var row *sql.Row
		args := strings.Split(match, "|")

		if len(args) == 1 {
			row = findStmt.QueryRow(args[0])
		} else {
			row = findBySetStmt.QueryRow(args[0], args[1])
		}

			response += buildSearchResponse(row, multi)
	}

	msr.matches = []string{}
	return response, err
}

func buildSearchResponse(row *sql.Row, multi bool) string {
	name, cost, text, multiverseID := "", "","",0

	err := row.Scan(&name, &cost, &text, &multiverseID)
	if err != nil {
		return "Card not found!\n"
	}

	if multi {
		return fmt.Sprintf("%s %s ```%s ```", name, formatCost(cost), text)
	}
	return fmt.Sprintf("%s %s ```%s ```", formatImageURL(multiverseID), formatCost(cost), text)
}

func formatImageURL(multiverseID int) string {
	return fmt.Sprintf("http://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=%d&type=card", multiverseID)
}

func formatCost(cost string) string {
	subBrackets := strings.NewReplacer("{", ":", "}", ":")
	fixSymbols := strings.NewReplacer("W", "WW", "U", "UU", "B", "BB", "G", "GG", "R", "RR")
	return fixSymbols.Replace(subBrackets.Replace(cost))
}
