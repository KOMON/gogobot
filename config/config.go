package config

import (
	"github.com/BurntSushi/toml"
)

type responder struct {
	Regexp    string
	Responses []string
}

type responders struct {
	Responders []responder
}

func PopulateResponders() responders {
	var rs responders
	if _, err := toml.DecodeFile("responders.toml", &rs); err != nil {
		panic(err)
	}
	return rs
}
