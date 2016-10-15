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

var MTGSTATS_REGEXP *regexp.Regexp = regexp.MustCompile("#\\[\\[((?:\\w+:\\s*\\w+,?\\s*)+)\\]\\]")

type MtgStatsResponder struct {
	re      *regexp.Regexp
	matches []string
}

func (msr *MtgStatsResponder) Match(s string) bool {
	msr.matches = msr.re.FindStringSubmatch(s)
	
	if msr.matches == nil {
		return false
	} else {
		return true
	}
}

func (msr MtgStatsResponder) Respond(rtm *slack.RTM, channelID string) bool {
	if msr.matches == nil {
		return true
	}

	rtm.SendMessage(rtm.NewOutgoingMessage(constructStatsResponse(msr.matches), channelID))
	return true
}

func constructStatsResponse(match []string) string {
	page := 0
	res := makeStatReq(match, page)

	if  res == nil {
		return "0"
	} else {
		var result []MtgCard

		body, _ := ioutil.ReadAll(res.Body)
		_ = json.Unmarshal(body, &result)

		total := len(result)
		res.Body.Close()

		for len(result) == 100 {
			page++
			res = makeStatReq(match, page)
			if res == nil {
				break
			} else {
				body, _ := ioutil.ReadAll(res.Body)
				_ = json.Unmarshal(body, &result)

				total += len(result)
				res.Body.Close()
			}
		}

		return fmt.Sprintf("%d", total)
	}
}

func makeStatReq(matches []string, page int) *http.Response {
	query := url.Values{}
	query.Set("page", fmt.Sprintf("%d",page))
	pairs := strings.Split(matches[1], ",")

	for _, pair := range pairs {
		pairray := strings.Split(pair, ":")
		query.Set(strings.TrimSpace(pairray[0]), strings.TrimSpace(pairray[1]))
	}
	url := DECKBREW_URL + query.Encode()
	res, _ := http.Get(url)
	return res
}


