package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/nlopes/slack"
)

type Handler interface {
	Match(string) bool
	Respond(*slack.RTM, string) bool
}

type ComplexHandler struct {
	re            *regexp.Regexp
	responses     []string
	buildResponse func([]string) (string, error)
	matches       []string
}

type Responder struct {
	re        *regexp.Regexp
	responses []string
	matches   []string
}

func (h *ComplexHandler) Match(s string) bool {
	h.matches = h.re.FindStringSubmatch(s)

	if h.matches == nil {
		return false
	} else {
		return true
	}
}

func (h ComplexHandler) Respond(rtm *slack.RTM, channelID string) bool {
	if h.matches == nil {
		return false
	}
	fmt.Printf("%t", h.responses)
	response, err := h.buildResponse(h.matches)
	fmt.Println(response)
	if err != nil {
		fmt.Println(err)
		return false
	}

	rtm.SendMessage(rtm.NewOutgoingMessage(response, channelID))
	return true
}

func (r *Responder) Match(s string) bool {
	r.matches = r.re.FindStringSubmatch(s)

	if r.matches == nil {
		return false
	} else {
		return true
	}
}

func (r Responder) Respond(rtm *slack.RTM, channelID string) bool {
	if r.matches == nil {
		return false
	}

	rand.Seed(time.Now().Unix())
	response := r.responses[rand.Int()%len(r.responses)]

	rtm.SendMessage(rtm.NewOutgoingMessage(response, channelID))
	return true
}
