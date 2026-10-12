package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/lucianaamm/goexpert-desafio6-Clima-por-CEP-no-Google-CloudRun/internal/handler"
	"github.com/lucianaamm/goexpert-desafio6-Clima-por-CEP-no-Google-CloudRun/internal/viacep"
	"github.com/lucianaamm/goexpert-desafio6-Clima-por-CEP-no-Google-CloudRun/internal/weatherapi"
)

func main() {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Fatal("WEATHER_API_KEY não definida")
	}

	// O Cloud Run injeta a porta na variável PORT.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	h := &handler.WeatherHandler{
		Locations: viacep.NewClient(httpClient),
		Weather:   weatherapi.NewClient(apiKey, httpClient),
	}

	mux := http.NewServeMux()
	mux.Handle("GET /weather/{cep}", h)
	mux.Handle("GET /{$}", h)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("servidor ouvindo na porta %s", port)
	log.Fatal(srv.ListenAndServe())
}
