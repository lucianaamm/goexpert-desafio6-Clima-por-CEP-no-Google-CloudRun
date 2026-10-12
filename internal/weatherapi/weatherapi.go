// Package weatherapi consulta a temperatura atual na WeatherAPI (weatherapi.com).
package weatherapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const DefaultBaseURL = "https://api.weatherapi.com/v1"

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(apiKey string, httpClient *http.Client) *Client {
	return &Client{BaseURL: DefaultBaseURL, APIKey: apiKey, HTTPClient: httpClient}
}

type response struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

// CurrentCelsius retorna a temperatura atual, em Celsius, da localização informada.
func (c *Client) CurrentCelsius(ctx context.Context, location string) (float64, error) {
	q := url.Values{}
	q.Set("key", c.APIKey)
	q.Set("q", location)
	q.Set("aqi", "no")
	endpoint := c.BaseURL + "/current.json?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		// *url.Error inclui a URL (com a chave da API); propaga só a causa.
		var uerr *url.Error
		if errors.As(err, &uerr) {
			err = uerr.Err
		}
		return 0, fmt.Errorf("weatherapi request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weatherapi: unexpected status %d", resp.StatusCode)
	}

	var body response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("weatherapi decode: %w", err)
	}
	return body.Current.TempC, nil
}
