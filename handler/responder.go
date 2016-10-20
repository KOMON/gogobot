package handler

import (
	"math/rand"
	"time"
	"regexp"
)

type Responder struct {
	RE        *regexp.Regexp
	Responses []string
	Matches   []string
}

func (r Responder) Respond() (string, error) {
	if r.Matches == nil {
		return "", &handlerError{"no matches available"}
	}

	rand.Seed(time.Now().Unix())
	response := r.Responses[rand.Int()%len(r.Responses)]

	return response, nil
}

func (r *Responder) Match(s string) bool {
	r.Matches = r.RE.FindStringSubmatch(s)

	if r.Matches == nil {
		return false
	}

	return true
}
