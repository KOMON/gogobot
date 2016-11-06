package handler

import (
	"fmt"
	"strings"
	"unicode"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	sq "github.com/Masterminds/squirrel"
)

type MtgStatsResponder struct {
	db *sql.DB
	matches []string
}

type Query map[string][]string

func NewMtgStats() MtgStatsResponder {
	db, err := sql.Open("sqlite3", "mtg.db")
	if err != nil {
		log.Fatal(err)
	}

	return MtgStatsResponder{
		db: db,
	}
}

func (msr *MtgStatsResponder) Match(s string) bool {
	msr.matches = []string{}

	for i:= strings.Index(s, "#[["); i!= -1; i = strings.Index(s, "#[[") {
		j := strings.Index(s[i+3:], "]]")
		msr.matches = append(msr.matches, s[i+3:j+3])
		s = s[j+2:]
	}

	if len(msr.matches) > 0 {
		return true
	}

	return false
}

func (msr MtgStatsResponder) Respond() (string, error) {
	query := Query{}
	args := strings.Split(msr.matches[0], ",")
	for _, arg := range args {
		kv := strings.Split(arg, ":")
		query[strings.TrimSpace(kv[0])] = strings.Split(strings.TrimSpace(kv[1]), "|")
	}
	return msr.runSearch(query)
}

func (msr MtgStatsResponder) runSearch(query Query) (string, error) {
	//#[[verb: avg(cmc)]]
	//verbs: avg, count, min, max, sum
	// avg - we just want the average
	// count - we just want the count
	// sum - we just want the sum
	// min - we want the min column as well as name and set and image?
	// max - we want the max column as well as ...

	var stmt sq.SelectBuilder

	if len(query["verb"]) != 0 {
		//select
		if v := strings.ToLower(query["verb"][0][0:2]); v == "min" || v == "max" {
			stmt = sq.Select("name", "set_code", "multiverse_id", query["verb"][0]).
				From("cards").
				Join("set_cards on cards.id=set_cards.id")
		} else if query["verb"][0] == ""{
			stmt = sq.Select("count(id)").From("cards")
		} else {
			stmt = sq.Select(query["verb"][0]).From("cards")
		}
	}

	for k,v := range query {
		switch k {
		case "verb":
			continue
		case "colors":
			stmt = stmt.Join("card_color on cards.id=card_color.id").
				Where(genColorQuery(v))
		case "colorID":
			stmt = stmt.Join("card_colorID on cards.id=card_colorID.id").
				Where(genColorQuery(v))
		case "supertypes":
			stmt = stmt.Join("card_supertype on cards.id=card_supertype.id").
				Where(sq.Eq{"supertype": v})
		case "types":
			stmt = stmt.Join("card_type on cards.id=card_type.id").
				Where(sq.Eq{"type": v})
		case "subtypes":
			stmt = stmt.Join("card_subtype on cards.id=card_subtype.id").
				Where(sq.Eq{"subtype": v})
		default:
			eq := sq.Eq{}
			eq[k] = v
			stmt = stmt.Where(eq)
		}
	}
	rows, err := stmt.RunWith(msr.db).Query()
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	return buildResponse(rows, query["verb"][0])
}

func buildResponse(rows *sql.Rows, verb string) (string, error) {
	if verb == "" || strings.ToLower(verb) == "count" {
		count := 0
		rows.Next()
		err := rows.Scan(&count)
		if err != nil {
			return "Error! ", err
		}

		return fmt.Sprintf("Count: %d", count), err
	} 
	if v := strings.ToLower(verb[0:2]); v == "min" || v == "max" {
		name, setCode, multiverseID, num := "", "", "", 0
		rows.Next()
		err := rows.Scan(&name, &setCode, &multiverseID, &num)
		if err != nil {
			return "Error! ", err
		}
		return fmt.Sprintf("%s: %d\nhttp://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=%s&type=card\n %s %s",
		verb, num, multiverseID, name, setCode), err
	}
	num := 0.0
	rows.Next()
	err := rows.Scan(&num)
	if err != nil {
		return "Error! ", err
	}
	return fmt.Sprintf("%s: %f", verb, num), err
}

func genColorQuery(colors []string) string {
	query := "0"
	for _, color := range colors {
		if len(color) > 1 {
			query += "|1"
			for _, char := range color {
				query += "&" + string(unicode.ToLower(char))
			}
		} else {
			query += "|" + strings.ToLower(color)
		}
	}
	return query
}
