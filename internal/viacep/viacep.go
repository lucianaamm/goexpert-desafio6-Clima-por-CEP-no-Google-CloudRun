// Package viacep consulta a localização de um CEP na API ViaCEP.
package viacep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const DefaultBaseURL = "https://viacep.com.br/ws"

// ErrNotFound indica que o CEP tem formato válido mas não existe na base.
var ErrNotFound = errors.New("zipcode not found")

// Location contém os dados do CEP usados pelo sistema.
type Location struct {
	City  string
	State string // nome completo do estado (ex.: "São Paulo")
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{BaseURL: DefaultBaseURL, HTTPClient: httpClient}
}

type response struct {
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Estado     string `json:"estado"`
	// A ViaCEP já retornou "erro" como string ("true") e como booleano (true).
	Erro any `json:"erro"`
}

// Lookup busca a cidade do CEP. Espera um CEP já validado (8 dígitos).
func (c *Client) Lookup(ctx context.Context, cep string) (Location, error) {
	endpoint := fmt.Sprintf("%s/%s/json/", c.BaseURL, url.PathEscape(cep))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Location{}, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Location{}, fmt.Errorf("viacep request: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound:
		return Location{}, ErrNotFound
	case resp.StatusCode != http.StatusOK:
		return Location{}, fmt.Errorf("viacep: unexpected status %d", resp.StatusCode)
	}

	var body response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Location{}, fmt.Errorf("viacep decode: %w", err)
	}
	if isTrue(body.Erro) || body.Localidade == "" {
		return Location{}, ErrNotFound
	}
	state := body.Estado
	if state == "" {
		state = body.UF
	}
	return Location{City: body.Localidade, State: state}, nil
}

func isTrue(v any) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return b == "true"
	}
	return false
}
