package H

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Rate struct {
	Currency  string `json:"currency"`
	Rate      string `json:"rate"`
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Open      string `json:"open"`
	Close     string `json:"close"`
	Timestamp string `json:"timestamp"`
}
type RateVes struct {
	Rate float64 `json:"rate"`
	Avg  float64 `json:"avg"`
	Time uint64  `json:"time"`
}

func FetchRates() ([]Rate, error) {
	// Define the URL
	url := "https://kijam.com/lic/rate/?no_cache=" + strconv.FormatInt(time.Now().UTC().Unix(), 10)

	// Perform the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var rates []Rate
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return nil, err
	}

	return rates, nil
}

func FetchVes() (*RateVes, error) {
	// Define the URL
	url := "https://kijam.com/lic/bcv/?no_cache=" + strconv.FormatInt(time.Now().UTC().Unix(), 10)

	// Perform the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var rate RateVes
	if err := json.NewDecoder(resp.Body).Decode(&rate); err != nil {
		return nil, err
	}

	return &rate, nil
}
