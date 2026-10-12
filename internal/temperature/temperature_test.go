package temperature

import "testing"

func TestCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		c, want float64
	}{
		{0, 32},
		{100, 212},
		{-40, -40},
		{28.5, 83.3},
		{37, 98.6},
	}
	for _, tt := range tests {
		if got := CelsiusToFahrenheit(tt.c); got != tt.want {
			t.Errorf("CelsiusToFahrenheit(%v) = %v, want %v", tt.c, got, tt.want)
		}
	}
}

func TestCelsiusToKelvin(t *testing.T) {
	tests := []struct {
		c, want float64
	}{
		{0, 273},
		{100, 373},
		{-273, 0},
		{28.5, 301.5},
	}
	for _, tt := range tests {
		if got := CelsiusToKelvin(tt.c); got != tt.want {
			t.Errorf("CelsiusToKelvin(%v) = %v, want %v", tt.c, got, tt.want)
		}
	}
}

func TestFromCelsius(t *testing.T) {
	got := FromCelsius(28.5)
	want := Temperatures{Celsius: 28.5, Fahrenheit: 83.3, Kelvin: 301.5}
	if got != want {
		t.Errorf("FromCelsius(28.5) = %+v, want %+v", got, want)
	}
}
