package handler

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/komon/gogobot/deckbrew"
)

var subBrackets *regexp.Regexp = regexp.MustCompile("[{}]")
var fixSymbols *regexp.Regexp = regexp.MustCompile(":([WURBGX]):")

type multiMTGError struct {
	errors []error
}

func (e *multiMTGError) Error() string {
	s := ""

	for i, e := range e.errors {
		if e != nil {
			s += fmt.Sprintf("%d: %s\n", i, e.Error())
		}
	}

	return s
}

func (e *multiMTGError) add(err error) {
	e.errors = append(e.errors, err)
}

type MtgSearchResponder struct {
		RE      *regexp.Regexp
		Matches [][]string
}


func (msr *MtgSearchResponder) Match(s string) bool {
	msr.Matches = msr.RE.FindAllStringSubmatch(s, -1)

	if msr.Matches == nil {
		return false
	}

	return true
}

func (msr MtgSearchResponder) Respond() (string, error) {
	if msr.Matches == nil {
		return "", &handlerError{ "MtgSearchResponder: no matches" }
	}

	response := ""
	errs := &multiMTGError{}
	multi := len(msr.Matches) > 1
	
	for _, match := range msr.Matches {

		query := url.Values{}
		query.Set("name", match[1])

		if match[2] != "" {
			query.Set("edition", match[2])
		}

		cards, err := deckbrew.GetCards(query)

		if err != nil || len(*cards) == 0 {
			response += "Card Not Found!\n"
			errs.add(err)
			continue
		}

		c, e := findCard(cards, match[1], match[2])
		s := buildSearchResponse(c, e, multi)

		response += s
	}

	return response, errs
}

func findCard(cards *[]deckbrew.Card, name string, ed string) (*deckbrew.Card, *deckbrew.Edition) {
	if cards == nil {
		return nil, nil
	}
	resCard := &(*cards)[0]

	for _, card := range *cards {
		if strings.EqualFold(card.Name, name) {
			resCard = &card
			break
		}
	}

	resEd := &(resCard.Editions[len(resCard.Editions)-1])
	if ed != "" {
		for _, e := range resCard.Editions {
			if strings.EqualFold(e.SetID, ed) {
				resEd = &e
				break
			}
		}
	}

	return resCard, resEd
}

func buildSearchResponse(c *deckbrew.Card, e *deckbrew.Edition, multi bool) string {
	if c == nil || e == nil {
		return "Card Not Found!"
	}
	format := "%s\n%s```%s```"
	cost := formatCost(c.Cost)

	if !multi {
		return fmt.Sprintf(format, e.ImageURL, cost, c.Text)
	}
	return fmt.Sprintf(format, c.Name, cost, c.Text)
}

func formatCost(cost string) string {
	cost = subBrackets.ReplaceAllString(cost, ":")
	return fixSymbols.ReplaceAllString(cost, ":$1$1:")
}
