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
		if v := strings.ToLower(query["verb"][0][0:3]); v == "min" || v == "max" {
			stmt = sq.Select("name", "multiverse_id", query["verb"][0]).
				From("cards")
		} else {
			stmt = sq.Select(query["verb"][0]).From("cards")
		}
	} else {
		stmt = sq.Select("count(cards.id)").From("cards")
		
	}
	for k,v := range query {
		switch k {
		case "verb":
			continue
		case "names", "name":
			eq, not := splitNegatives(v)
			fmt.Println(eq, not)
			if len(eq[0]) != 0 {
				stmt = stmt.Where(sq.Eq{"cards.name": eq})
			}
			if len(not[0]) != 0 {
				stmt = stmt.Where(sq.NotEq{"cards.name": not})
			}
		case "colors", "color":
			stmt = stmt.Join("card_color on cards.id=card_color.id").
				Where(genColorQuery(v, false))
		case "colorIDs", "colorID":
			stmt = stmt.Join("card_colorID on cards.id=card_colorID.id").
				Where(genColorQuery(v, true))
		case "supertypes", "supertype":
			stmt = msr.joinAndWhere("card_supertype", "supertype", v, stmt, strings.Title)
		case "types", "type":
			stmt = msr.joinAndWhere("card_type", "type", v, stmt, strings.Title)
		case "subtypes", "subtype":
			stmt = msr.joinAndWhere("card_subtype", "subtype", v, stmt, strings.Title)
		case "sets", "set", "set_codes", "set_code":
			stmt = msr.joinAndWhere("set_card", "set_code", v, stmt, strings.ToUpper)
		default:
			eq := sq.Eq{}
			eq[k] = v
			stmt = stmt.Where(eq)
		}
	}
	rows, err := stmt.Where("multiverse_id != 0").RunWith(msr.db).Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	return buildResponse(rows, query["verb"])
}

func (msr MtgStatsResponder) joinAndWhere(table string, column string, 
values []string, stmt sq.SelectBuilder, f func(string) string) sq.SelectBuilder {

	eq, not := splitNegatives(values)
	newStmt := stmt.Join(table + " on cards.id=" + table + ".id")
	disqualIDs := []string{}
	disqualQuery := sq.
		Select("cards.id").
		From("cards").
		Join(table + " on cards.id=" + table + ".id").
		Where(sq.Eq{table+"."+column: strMap(not, f)})

	disqualRows, err := disqualQuery.RunWith(msr.db).Query()

	if err != nil {
		log.Print(err)
	} else {
		for disqualRows.Next() {
			ID := ""
			err := disqualRows.Scan(&ID)

			if err != nil {
				log.Print(err)
			}

			disqualIDs = append(disqualIDs, ID)
		}
	}
	if len(eq[0]) != 0 {
		newStmt = newStmt.Where(sq.Eq{table+"."+column: strMap(eq, f)})
	}
	if len(disqualIDs) != 0 {
		newStmt = newStmt.Where(sq.NotEq{"cards.id": disqualIDs})
	}
	return newStmt
}


func buildResponse(rows *sql.Rows, verb []string) (string, error) {
	if len(verb) == 0 || strings.ToLower(verb[0]) == "count" {
		count := 0
		rows.Next()
		err := rows.Scan(&count)
		if err != nil {
			return "Error! ", err
		}

		return fmt.Sprintf("Count: %d", count), err
	}
	if v := strings.TrimSpace(strings.ToLower(verb[0][0:3])); v == "min" || v == "max" {
		name, multiverseID, num := "", "", 0
		rows.Next()
		err := rows.Scan(&name, &multiverseID, &num)
		if err != nil {
			return "Error! ", err
		}
		return fmt.Sprintf("%s: %d\nhttp://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=%s&type=card\n %s",
		verb[0], num, multiverseID, name), err
	}
	num := 0.0
	rows.Next()
	err := rows.Scan(&num)
	if err != nil {
		return "Error! ", err
	}
	return fmt.Sprintf("%s: %f", verb[0], num), err
}

func genColorQuery(colors []string, ID bool) string {
	query := "0"
	for _, color := range colors {
		if len(color) > 1 {
			query += "|1"
			for _, char := range color {
				if ID {
					query += "&" + "card_colorID." + string(unicode.ToLower(char))
				} else {
					query += "&" + "card_color." + string(unicode.ToLower(char))
				}
			}
		} else {
			if ID {
				query += "|" + "card_colorID." + strings.ToLower(color)				
			} else {
				query += "|" + "card_color." + strings.ToLower(color)
			}
		}
	}
	return strings.ToUpper(query)
}

func strMap(ss []string, f func(string) string) []string {
	mapped := make([]string, len(ss))
	for i, s := range ss {
		mapped[i] = f(s)
	}

	return mapped
}

func strFilter(ss []string, f func(string) bool) []string {
	matches := make([]string, len(ss))
	for i, s := range ss {
		if f(s) {
			matches[i] = s
		}
	}
	return matches
}

func splitNegatives(ss []string) ([]string, []string) {
	matches := strFilter(ss, func(s string) bool {
		if s[0] == '!'{
			return false
		}
		return true;
	})

	non := strFilter(ss, func(s string) bool {
		if len(s) != 0 && s[0] == '!' {
			return true
		}
		return false
	})
	if len(non[0]) != 0 {
		non = strMap(non, func(s string) string {
			return s[1:]
		})
	}
	return matches, non
}
