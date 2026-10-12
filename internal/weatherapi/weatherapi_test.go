package weatherapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCurrentCelsius(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/current.json" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("key"); got != "secret" {
			t.Errorf("key = %q", got)
		}
		if got := r.URL.Query().Get("q"); got != "São Paulo, São Paulo, Brazil" {
			t.Errorf("q = %q", got)
		}
		w.Write([]byte(`{"current":{"temp_c":28.5}}`))
	}))
	defer srv.Close()

	c := NewClient("secret", srv.Client())
	c.BaseURL = srv.URL

	got, err := c.CurrentCelsius(context.Background(), "São Paulo, São Paulo, Brazil")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 28.5 {
		t.Errorf("got %v, want 28.5", got)
	}
}

func TestCurrentCelsiusUpstreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewClient("secret", srv.Client())
	c.BaseURL = srv.URL

	_, err := c.CurrentCelsius(context.Background(), "Recife")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Errorf("error leaks API key: %v", err)
	}
}

func TestCurrentCelsiusConnectionErrorDoesNotLeakKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // servidor fechado força erro de conexão

	c := NewClient("secret", http.DefaultClient)
	c.BaseURL = srv.URL

	_, err := c.CurrentCelsius(context.Background(), "Recife")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Errorf("error leaks API key: %v", err)
	}
}
