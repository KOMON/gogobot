package handler

import (
	"fmt"
	"strings"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	sq "github.com/Masterminds/squirrel"
)


type MtgSearchResponder struct {
	db *sql.DB
	matches []string
}

func NewMtgSearch() MtgSearchResponder {
	db, err := sql.Open("sqlite3", "mtg.db")
	if err != nil {
		log.Fatal(err)
	}
	
	return MtgSearchResponder{
		db: db,
	}
}

func (msr *MtgSearchResponder) Match(s string) bool {
	msr.matches = []string{}

	for i:= strings.Index(s, "[["); i != -1; i = strings.Index(s, "[[") {
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
	var err error
	response := "" 
	multi := len(msr.matches) > 1
	for _, match := range msr.matches {
		var rows *sql.Rows
		
		args := strings.Split(match, "|")

		if len(args) == 1 {
			rows, err = msr.runSearch(args[0], "")
		} else {
			rows, err = msr.runSearch(args[0], args[1])
		}

		if err != nil {
			response += "Card Not Found!\n"
			log.Println(err)
			continue
		}

		defer rows.Close()
		response += buildSearchResponse(rows, args[0], multi)
	}

	msr.matches = []string{}
	return response, err
}

func (msr MtgSearchResponder) runSearch(qName string, qSet string) (*sql.Rows, error) {
	nameQuery := sq.
		Select("id").
		From("virt_cards").
		Where("name match ? and multiverse_id != 0", qName)

	virtRows, err := nameQuery.RunWith(msr.db).Query()

	if err != nil {
		log.Fatal(err)
	}

	defer virtRows.Close()
	IDs := []string{}

	for virtRows.Next() {
		id := ""

		err := virtRows.Scan(&id)

		if err != nil {
			log.Fatal(err)
		}

		IDs = append(IDs, id)
	}

	var query sq.SelectBuilder
	if qSet != "" {
		query = sq.
			Select("name","mana_cost", "card_text", "multiverse_id").
			From("cards").
			Join("set_card on cards.id = set_card.id").
			Where(sq.Eq{"cards.id": IDs, "set_code": strings.ToUpper(qSet)})
	} else {
		query = sq.
			Select("name", "mana_cost", "card_text", "multiverse_id").
			From("cards").
			Where(sq.Eq{"id": IDs}).
			OrderBy("name")
	}
	return query.RunWith(msr.db).Query()
}

func buildSearchResponse(rows *sql.Rows, qName string, multi bool) string {
	name, cost, text, multiverseID := "", "","",0

	if rows == nil || !rows.Next() {
		return "Card not found\n"
	}
	
	err := rows.Scan(&name, &cost, &text, &multiverseID)
	if err != nil {
		return "Card not found!\n"
	}

	if !strings.EqualFold(name, qName) {
		for rows.Next() {
			forName, forCost, forText, forMultiverseID := "", "", "", 0
			err := rows.Scan(&forName, &forCost, &forText, &forMultiverseID)

			if err != nil {
				break
			}
			if strings.EqualFold(forName, qName) {
				name, cost, text, multiverseID = forName, forCost, forText, forMultiverseID
				break
			}
		}
	}

	if multi {
		return fmt.Sprintf("%s %s ```%s ```\n", name, formatCost(cost), text)
	}
	return fmt.Sprintf("%s %s ```%s ```", formatImageURL(multiverseID), formatCost(cost), text)
}

func formatImageURL(multiverseID int) string {
	return fmt.Sprintf("http://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=%d&type=card", multiverseID)
}

func formatCost(cost string) string {
	subBrackets := strings.NewReplacer("{", ":", "}", ":")
	fixSymbols := strings.NewReplacer("W", "ww", "U", "uu", "B", "bb", "G", "gg", "R", "rr")
	return fixSymbols.Replace(subBrackets.Replace(cost))
}
