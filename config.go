package main

import (
	"regexp"

	"github.com/BurntSushi/toml"
)

type responder struct {
	Regexp    string
	Responses []string
}

type responders struct {
	Responders []responder
}

func populateResponders() []Handler {
	var rs responders

	handlers := []Handler{}

	if _, err := toml.DecodeFile("responders.toml", &rs); err != nil {
		panic(err)
	}

	for _, r := range rs.Responders {
		re, err := regexp.Compile(r.Regexp)

		if err != nil {
			panic(err)
		}

		new := &Responder{
			re:        re,
			responses: r.Responses,
			matches:   []string{},
		}
		handlers = append(handlers, new)
	}
	return handlers
}
