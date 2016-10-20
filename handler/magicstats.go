package handler

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/komon/gogobot/deckbrew"
)

type MtgStatsResponder struct {
	RE *regexp.Regexp
	Matches []string
}

func (msr *MtgStatsResponder) Match(s string) bool {
	msr.Matches = msr.RE.FindStringSubmatch(s)

	if msr.Matches == nil {
		return false
	}

	return true
}

func (msr MtgStatsResponder) Respond() (string, error) {
	if msr.Matches == nil {
		return "", &handlerError{ "MtgStatsResponder: no matches" }
	}

	query := url.Values{}
	query.Set("page", "0")

	for _, pair := range strings.Split(msr.Matches[1], ",") {
		pairray := strings.Split(pair, ":")
		query.Add(strings.TrimSpace(pairray[0]), strings.TrimSpace(pairray[1]))
	}

	return buildStatsResponse(query)
}

func buildStatsResponse(query url.Values) (string, error) {
	total, page := 0, 0
	
	cards, err := deckbrew.GetCards(query)

	if err != nil || cards == nil {
		return "0", err
	}
	
	total += len(*cards)

	for len(*cards) == 100 {
		page++
		query.Set("page", fmt.Sprintf("%d", page))
		cards, err = deckbrew.GetCards(query)

		if err != nil || cards == nil {
			break
		}
		total += len(*cards)
	}

	return fmt.Sprintf("%d", total), err
}
