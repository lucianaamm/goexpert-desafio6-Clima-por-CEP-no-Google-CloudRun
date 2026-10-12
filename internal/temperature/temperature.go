// Package temperature converte temperaturas a partir de graus Celsius.
package temperature

import "math"

// Temperatures representa a resposta de sucesso da API.
type Temperatures struct {
	Celsius    float64 `json:"temp_C"`
	Fahrenheit float64 `json:"temp_F"`
	Kelvin     float64 `json:"temp_K"`
}

// CelsiusToFahrenheit aplica F = C * 1.8 + 32.
func CelsiusToFahrenheit(c float64) float64 {
	return round(c*1.8 + 32)
}

// CelsiusToKelvin aplica K = C + 273.
func CelsiusToKelvin(c float64) float64 {
	return round(c + 273)
}

// FromCelsius monta as três escalas a partir de uma temperatura em Celsius.
func FromCelsius(c float64) Temperatures {
	return Temperatures{
		Celsius:    round(c),
		Fahrenheit: CelsiusToFahrenheit(c),
		Kelvin:     CelsiusToKelvin(c),
	}
}

// round arredonda para 2 casas decimais, eliminando ruído de ponto flutuante.
func round(v float64) float64 {
	return math.Round(v*100) / 100
}
