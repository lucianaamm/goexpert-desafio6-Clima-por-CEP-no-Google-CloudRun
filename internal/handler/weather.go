// Package handler expõe o endpoint HTTP de clima por CEP.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	"github.com/lucianaamm/goexpert-desafio6-Clima-por-CEP-no-Google-CloudRun/internal/temperature"
	"github.com/lucianaamm/goexpert-desafio6-Clima-por-CEP-no-Google-CloudRun/internal/viacep"
)

const (
	msgInvalidZipcode  = "invalid zipcode"
	msgZipcodeNotFound = "can not find zipcode"
	msgInternalError   = "internal server error"
)

var cepPattern = regexp.MustCompile(`^\d{8}$`)

// LocationFinder resolve um CEP em uma localização.
type LocationFinder interface {
	Lookup(ctx context.Context, cep string) (viacep.Location, error)
}

// WeatherProvider retorna a temperatura atual (Celsius) de uma localização.
type WeatherProvider interface {
	CurrentCelsius(ctx context.Context, location string) (float64, error)
}

type WeatherHandler struct {
	Locations LocationFinder
	Weather   WeatherProvider
}

// ServeHTTP atende GET /weather/{cep} e GET /?cep={cep}.
func (h *WeatherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cep := r.PathValue("cep")
	if cep == "" {
		cep = r.URL.Query().Get("cep")
	}

	if !IsValidCEP(cep) {
		writeError(w, http.StatusUnprocessableEntity, msgInvalidZipcode)
		return
	}

	loc, err := h.Locations.Lookup(r.Context(), cep)
	if errors.Is(err, viacep.ErrNotFound) {
		writeError(w, http.StatusNotFound, msgZipcodeNotFound)
		return
	}
	if err != nil {
		log.Printf("lookup cep %s: %v", cep, err)
		writeError(w, http.StatusInternalServerError, msgInternalError)
		return
	}

	celsius, err := h.Weather.CurrentCelsius(r.Context(), loc.City+", "+loc.State+", Brazil")
	if err != nil {
		log.Printf("weather for %s/%s: %v", loc.City, loc.State, err)
		writeError(w, http.StatusInternalServerError, msgInternalError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(weatherResponse{
		City:         loc.City,
		State:        loc.State,
		Temperatures: temperature.FromCelsius(celsius),
	})
}

// weatherResponse é o corpo de sucesso: localização seguida das temperaturas.
type weatherResponse struct {
	City  string `json:"city"`
	State string `json:"state"`
	temperature.Temperatures
}

// IsValidCEP aceita exatamente 8 dígitos, sem hífen ou outros caracteres.
func IsValidCEP(cep string) bool {
	return cepPattern.MatchString(cep)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	w.Write([]byte(msg))
}
