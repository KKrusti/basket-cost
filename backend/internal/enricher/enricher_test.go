package enricher

import (
	"testing"
)

// ---------- normalise ----------

func TestNormalise(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase", "LECHE ENTERA", "leche entera"},
		{"strip accents", "ARRÒS INTEGRAL", "arros integral"},
		{"strip accents es", "ACEITE DE OLIVA VIRGEN EXTRA", "aceite de oliva virgen extra"},
		{"collapse spaces", "  PAN  DE  MOLDE  ", "pan de molde"},
		{"ñ to n", "PIÑONES", "pinones"},
		{"ç to c", "MOZZARELLA FRESCA", "mozzarella fresca"},
		{"strip punctuation", "ENERGY DRINK (KATRINE)", "energy drink katrine"},
		{"numbers preserved", "LECHE 1L", "leche 1l"},
		{"mixed case accented", "Làmina d'Embolicar", "lamina d embolicar"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalise(tt.input)
			if got != tt.want {
				t.Errorf("normalise(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------- deaccent ----------

func TestDeaccent_SpanishVowels(t *testing.T) {
	tests := []struct {
		in   rune
		want rune
	}{
		{'á', 'a'}, {'é', 'e'}, {'í', 'i'}, {'ó', 'o'}, {'ú', 'u'},
		{'à', 'a'}, {'è', 'e'}, {'ì', 'i'}, {'ò', 'o'}, {'ù', 'u'},
		{'Á', 'a'}, {'É', 'e'}, {'Í', 'i'}, {'Ó', 'o'}, {'Ú', 'u'},
		{'ñ', 'n'}, {'Ñ', 'n'}, {'ç', 'c'}, {'Ç', 'c'},
		{'a', 'a'}, {'z', 'z'}, {'0', '0'}, // pass-through
	}
	for _, tt := range tests {
		got := deaccent(tt.in)
		if got != tt.want {
			t.Errorf("deaccent(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
