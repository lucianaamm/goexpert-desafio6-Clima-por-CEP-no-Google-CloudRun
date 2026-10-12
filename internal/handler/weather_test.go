package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucianaamm/goexpert-desafio6-Clima-por-CEP-no-Google-CloudRun/internal/temperature"
	"github.com/lucianaamm/goexpert-desafio6-Clima-por-CEP-no-Google-CloudRun/internal/viacep"
)

type fakeLocations struct {
	loc viacep.Location
	err error
}

func (f fakeLocations) Lookup(context.Context, string) (viacep.Location, error) {
	return f.loc, f.err
}

type fakeWeather struct {
	celsius  float64
	err      error
	gotQuery *string
}

func (f fakeWeather) CurrentCelsius(_ context.Context, location string) (float64, error) {
	if f.gotQuery != nil {
		*f.gotQuery = location
	}
	return f.celsius, f.err
}

func newMux(h *WeatherHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /weather/{cep}", h)
	mux.Handle("GET /{$}", h)
	return mux
}

func do(t *testing.T, h *WeatherHandler, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	newMux(h).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestSuccess(t *testing.T) {
	var query string
	h := &WeatherHandler{
		Locations: fakeLocations{loc: viacep.Location{City: "São Paulo", State: "São Paulo"}},
		Weather:   fakeWeather{celsius: 28.5, gotQuery: &query},
	}

	for _, target := range []string{"/weather/01001000", "/?cep=01001000"} {
		t.Run(target, func(t *testing.T) {
			rec := do(t, h, target)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q", ct)
			}
			var got weatherResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			want := weatherResponse{
				City:         "São Paulo",
				State:        "São Paulo",
				Temperatures: temperature.Temperatures{Celsius: 28.5, Fahrenheit: 83.3, Kelvin: 301.5},
			}
			if got != want {
				t.Errorf("got %+v, want %+v", got, want)
			}
			if query != "São Paulo, São Paulo, Brazil" {
				t.Errorf("weather query = %q", query)
			}
		})
	}
}

func TestResponseFieldNames(t *testing.T) {
	h := &WeatherHandler{
		Locations: fakeLocations{loc: viacep.Location{City: "Recife", State: "PE"}},
		Weather:   fakeWeather{celsius: 30},
	}
	rec := do(t, h, "/weather/50030230")

	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"city", "state", "temp_C", "temp_F", "temp_K"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing field %q in %s", key, rec.Body)
		}
	}
}

func TestInvalidZipcode(t *testing.T) {
	h := &WeatherHandler{
		Locations: fakeLocations{err: errors.New("must not be called")},
		Weather:   fakeWeather{err: errors.New("must not be called")},
	}
	for _, target := range []string{
		"/weather/1234567",   // 7 dígitos
		"/weather/123456789", // 9 dígitos
		"/weather/0100100a",  // letra
		"/weather/01001-000", // hífen
		"/?cep=",             // vazio
		"/",                  // ausente
	} {
		t.Run(target, func(t *testing.T) {
			rec := do(t, h, target)
			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("status = %d, want 422", rec.Code)
			}
			if rec.Body.String() != "invalid zipcode" {
				t.Errorf("body = %q", rec.Body)
			}
		})
	}
}

func TestZipcodeNotFound(t *testing.T) {
	h := &WeatherHandler{
		Locations: fakeLocations{err: viacep.ErrNotFound},
		Weather:   fakeWeather{},
	}
	rec := do(t, h, "/weather/99999999")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	if rec.Body.String() != "can not find zipcode" {
		t.Errorf("body = %q", rec.Body)
	}
}

func TestUpstreamFailures(t *testing.T) {
	tests := map[string]*WeatherHandler{
		"viacep falha": {
			Locations: fakeLocations{err: errors.New("boom")},
			Weather:   fakeWeather{},
		},
		"weatherapi falha": {
			Locations: fakeLocations{loc: viacep.Location{City: "Recife", State: "PE"}},
			Weather:   fakeWeather{err: errors.New("boom")},
		},
	}
	for name, h := range tests {
		t.Run(name, func(t *testing.T) {
			rec := do(t, h, "/weather/50030230")
			if rec.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want 500", rec.Code)
			}
		})
	}
}
