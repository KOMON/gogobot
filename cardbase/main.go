package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"fmt"
	"strconv"
)

const createSets string = `create table sets (
name varchar(50),
code varchar(4),
release_date varchar(10),
type varchar(10),
block varchar(50)
)`

const createCards string = `create table cards (
id varchar(40),
layout varchar(10),
name varchar(255),
mana_cost varchar(10),
cmc integer,
type varchar(100),
card_text text,
flavor text,
artist varchar(50),
number varchar(20),
power integer,
toughness integer,
loyalty integer,
multiverse_id integer,
timeshifted boolean,
reserved boolean,
release_date varchar(10),
mci_number varchar(4)
)`

const createSetCard string = `create table set_card (set_code varchar(4), id varchar(40))`
const createCardColor string = `create table card_color (
id varchar(40), r boolean, g boolean, u boolean, b boolean, w boolean, colorless boolean)`
const createCardColorID string = `create table card_colorID (
id varchar(40), r boolean, g boolean, u boolean, b boolean, w boolean, colorless boolean)`
const createCardSupertype string = `create table card_supertype(id varchar(40), supertype varchar(10))`
const createCardSubtype string = `create table card_subtype(id varchar(40), subtype varchar(20))`
const createCardType string = `create table card_type(id varchar(40), type varchar(20))`
const createCardRarity string = `create table card_rarity(id varchar(40), rarity varchar(12))`

var createTables []string = []string{
	createSets,
	createCards,
	createSetCard,
	createCardColor,
	createCardColorID,
	createCardSupertype,
	createCardSubtype,
	createCardType,
	createCardRarity,
}

func main() {
	os.Remove("./mtg.db")

	db, err := sql.Open("sqlite3", "./mtg.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, stmt := range createTables {
		_, err = db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	sets, err := loadSets("AllSets.json")
	if err != nil {
		log.Fatal(err)
	}
	setL := len(sets)
	setPos := 1
	for _, s := range sets {
		fmt.Printf("Importing set: %s %d/%d\n", s.Code, setPos, setL)
		setPos++
		ImportSet(db, s)
		cardL := len(s.Cards)
		for j, c := range s.Cards {
			fmt.Printf("Importing card %d/%d\r",j+1,cardL )
			if s.Type == "promo" {
				ImportCard(db, c, c.ReleaseDate)
			} else {
				ImportCard(db, c, s.ReleaseDate)
			}
			ImportSetCard(db, s, c)
			ImportCardColor(db, c)
			ImportCardColorID(db, c)
			ImportCardSupertype(db, c)
			ImportCardType(db, c)
			ImportCardSubtype(db, c)
			ImportCardRarity(db, c)
		}
	}
}

func ImportCard(db *sql.DB, c Card, releaseDate string) {
	const insertCard string = `insert into cards(
id, name, mana_cost, cmc, type, card_text,
flavor, artist, number, power, toughness,
loyalty, multiverse_id, timeshifted,
reserved, release_date, mci_number
) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(insertCard)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	p, err := strconv.ParseInt(c.Power,0,0)
	if err != nil {
		p = 0
	}
	t, err := strconv.ParseInt(c.Toughness,0,0)
	if err != nil {
		t = 0
	}
	_, err = stmt.Exec(c.ID, c.Name, c.ManaCost, c.CMC, c.Type, c.Text, c.Flavor,
		c.Artist, c.Number, p, t, c.Loyalty, c.MultiverseID,
		c.Timeshifted, c.Reserved, releaseDate, c.MCINumber)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

func ImportSet(db *sql.DB, s Set) {
	const insertSet string = `insert into sets(name, code, release_date, type, block) values(?,?,?,?,?)`
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(insertSet)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(s.Name, s.Code, s.ReleaseDate, s.Type, s.Block)

	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}

func ImportSetCard(db *sql.DB, s Set, c Card) {
	const insertSetCard string = `insert into set_card(set_code, id) values(?,?)`

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(insertSetCard)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(s.Code, c.ID)

	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}

func ImportCardColor(db *sql.DB, c Card) {
	const insertCardColor string = `insert into card_color(id, r, g, u, b, w, colorless) values(?,?,?,?,?,?,?)`
	r, g, u, b, w := false, false, false, false, false
	colorless := true
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(insertCardColor)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	if c.Colors != nil {
		colorless = false
		for _, color := range c.Colors {
			switch color {
			case "Red":
				r = true
			case "Blue":
				u = true
			case "Green":
				g = true
			case "Black":
				b = true
			case "White":
				w = true
			}
		}
	}
		_, err = stmt.Exec(c.ID, r, g, u, b, w, colorless)
		if err != nil {
			log.Fatal(err)
		}

	tx.Commit()
}

func ImportCardColorID(db *sql.DB, c Card) {
	const insertCardColorID string = `insert into card_colorID(id, r, g, u, b, w, colorless) values(?,?,?,?,?,?,?)`
	r, g, u, b, w := false, false, false, false, false
	colorless := true
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(insertCardColorID)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	if c.Colors != nil {
		colorless = false
		for _, color := range c.Colors {
			switch color {
			case "Red":
				r = true
			case "Blue":
				u = true
			case "Green":
				g = true
			case "Black":
				b = true
			case "White":
				w = true
			}
		}
	}
		_, err = stmt.Exec(c.ID, r, g, u, b, w, colorless)
		if err != nil {
			log.Fatal(err)
		}

	tx.Commit()

}

func ImportCardSupertype(db *sql.DB, c Card) {
	const insertCardSupertype string = `insert into card_supertype(id, supertype) values(?,?)`

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(insertCardSupertype)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, s := range c.Supertypes {
		_, err = stmt.Exec(c.ID, s)

		if err != nil {
			log.Fatal(err)
		}
	}

	tx.Commit()
}

func ImportCardType(db *sql.DB, c Card) {
	const insertCardType string = `insert into card_type(id, type) values(?,?)`

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(insertCardType)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, t := range c.Types {
		_, err = stmt.Exec(c.ID, t)

		if err != nil {
			log.Fatal(err)
		}
	}

	tx.Commit()
}

func ImportCardSubtype(db *sql.DB, c Card) {
	const insertCardSubtype string = `insert into card_subtype(id, subtype) values(?,?)`

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(insertCardSubtype)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, st := range c.Subtypes {
		_, err = stmt.Exec(c.ID, st)

		if err != nil {
			log.Fatal(err)
		}
	}

	tx.Commit()
}

func ImportCardRarity(db *sql.DB, c Card) {
	const insertCardRarity string = `insert into card_rarity(id, rarity) values(?,?)`

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(insertCardRarity)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, r := range c.Rarity.Rarities {
		_, err = stmt.Exec(c.ID, r)

		if err != nil {
			log.Fatal(err)
		}
	}

	tx.Commit()
}
