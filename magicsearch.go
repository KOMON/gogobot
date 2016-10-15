package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
)

var MTGSEARCH_REGEXP *regexp.Regexp = regexp.MustCompile("(?:[^#]\\[\\[(.*?)(\\|...)?\\]\\])+")
var SUB_BRACKETS_REGEXP *regexp.Regexp = regexp.MustCompile("[{}]")
var FIX_SYMBOLS_REGEXP *regexp.Regexp = regexp.MustCompile(":([WURBGX]):")

const DECKBREW_URL string = "https://api.deckbrew.com/mtg/cards?"

type MtgSearchResponder struct {
	re      *regexp.Regexp
	matches [][]string
}

type MtgCard struct {
	Name     string
	Cost     string
	Text     string
	Id       string
	Editions []MtgEditionsField
}

type MtgEditionsField struct {
	Multiverse_id float64
	Number        string
	Set_id        string
	Image_url     string
}

func (msr *MtgSearchResponder) Match(s string) bool {
	msr.matches = msr.re.FindAllStringSubmatch(s, -1)

	if msr.matches == nil {
		return false
	} else {
		return true
	}
}

func (msr MtgSearchResponder) Respond(rtm *slack.RTM, channelID string) bool {
	if msr.matches == nil {
		return true
	}

	if len(msr.matches) > 1 {
		for _, match := range msr.matches {
			rtm.SendMessage(rtm.NewOutgoingMessage(constructResponse(match, true), channelID))
		}
	} else {
		rtm.SendMessage(rtm.NewOutgoingMessage(constructResponse(msr.matches[0], false), channelID))
	}

	return true

}

func constructResponse(match []string, multi bool) string {
	var response string
	res := makeMTGReq(match)

	if res == nil {
		response = "Card Not Found"
	} else {
		defer res.Body.Close()
		var result []MtgCard
		body, _ := ioutil.ReadAll(res.Body)
		_ = json.Unmarshal(body, &result)

		if len(result) == 0 {
			response = "Card Not Found"
		} else {
			response = buildResponse(findCard(&result, match), match, multi)
		}
	}
	return response
}

func makeMTGReq(match []string) *http.Response {
	if match == nil {
		return nil
	}

	query := url.Values{}

	query.Set("name", match[1])

	if match[2] != "" {
		query.Set("edition", match[2])
	}

	url := DECKBREW_URL + query.Encode()
	res, _ := http.Get(url)

	return res
}

func findCard(cards *[]MtgCard, match []string) *MtgCard {
	for _, card := range *cards {
		if strings.EqualFold(card.Name, match[1]) {
			return &card
		}
	}
	return &(*cards)[0]
}

func buildResponse(card *MtgCard, match []string, multi bool) string {

	ed := card.Editions[len(card.Editions)-1]

	if match[2] != "" {
		for _, e := range card.Editions {
			if strings.EqualFold(e.Set_id, match[2][1:]) {
				ed = e
			}
		}
	}
	if !multi {
		return fmt.Sprintf("%s\n%s```%s```", ed.Image_url, formatCost(card.Cost), card.Text)
	} else {
		return fmt.Sprintf("%s\n%s```%s```", card.Name, formatCost(card.Cost), card.Text)
	}

}

func formatCost(cost string) string {
	cost = SUB_BRACKETS_REGEXP.ReplaceAllString(cost, ":")
	return FIX_SYMBOLS_REGEXP.ReplaceAllString(cost, ":$1$1:")
}
