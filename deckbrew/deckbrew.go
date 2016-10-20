package deckbrew

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

const deckbrewURL string = "https://api.deckbrew.com/mtg/cards?"

type (
	Card struct {
		Name     string
		Cost     string
		Text     string
		ID       string
		Editions []Edition
	}

	Edition struct {
		MultiverseID float64
		Number       string
		SetID        string `json:"set_id"`
		ImageURL     string `json:"image_url"`
	}
)

func GetCards(req url.Values) (*[]Card, error) {
	res, err := http.Get(deckbrewURL + req.Encode())
	if res == nil || err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result []Card

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &result)

	if err != nil {
		return nil, err
	}

	return &result, err
}

