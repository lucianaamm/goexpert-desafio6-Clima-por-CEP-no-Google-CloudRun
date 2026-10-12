package viacep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient(srv.Client())
	c.BaseURL = srv.URL
	return c
}

func TestLookupSuccess(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/01001000/json/" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Write([]byte(`{"cep":"01001-000","localidade":"São Paulo","uf":"SP","estado":"São Paulo"}`))
	})

	loc, err := c.Lookup(context.Background(), "01001000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc.City != "São Paulo" || loc.State != "São Paulo" {
		t.Errorf("got %+v", loc)
	}
}

func TestLookupFallsBackToUF(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"localidade":"Recife","uf":"PE"}`))
	})
	loc, err := c.Lookup(context.Background(), "50030230")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc.State != "PE" {
		t.Errorf("State = %q, want PE", loc.State)
	}
}

func TestLookupNotFound(t *testing.T) {
	bodies := map[string]string{
		"erro booleano": `{"erro": true}`,
		"erro string":   `{"erro": "true"}`,
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(body))
			})
			_, err := c.Lookup(context.Background(), "99999999")
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}
		})
	}
}

func TestLookupBadRequest(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	_, err := c.Lookup(context.Background(), "00000000")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLookupUpstreamError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := c.Lookup(context.Background(), "01001000")
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Errorf("expected generic error, got %v", err)
	}
}
