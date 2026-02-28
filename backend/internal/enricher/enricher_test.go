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

// ---------- keywords ----------

func TestKeywords_FiltersStopWords(t *testing.T) {
	// "de", "el", "la" are stop words; "oliva", "virgen" are not.
	got := keywords("aceite de oliva virgen")
	want := []string{"aceite", "oliva", "virgen"}
	if len(got) != len(want) {
		t.Fatalf("keywords(%q) = %v, want %v", "aceite de oliva virgen", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("keywords[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestKeywords_FiltersShortTokens(t *testing.T) {
	// "1l", "g" are too short (< 3 runes) — they should be dropped.
	got := keywords("leche 1l")
	if len(got) != 1 || got[0] != "leche" {
		t.Errorf("keywords(%q) = %v, want [leche]", "leche 1l", got)
	}
}

func TestKeywords_EmptyInput(t *testing.T) {
	got := keywords("")
	if len(got) != 0 {
		t.Errorf("keywords(%q) = %v, want []", "", got)
	}
}

// ---------- bestMatch ----------

func TestBestMatch_ExactMatch(t *testing.T) {
	index := ProductIndex{
		{Thumbnail: "https://example.com/leche.jpg", Keywords: []string{"leche", "entera"}},
		{Thumbnail: "https://example.com/pan.jpg", Keywords: []string{"pan", "integral", "molde"}},
	}
	// "leche entera" → localKW = [leche, entera]; must match first entry at score 1.0.
	url, ok := bestMatch([]string{"leche", "entera"}, index)
	if !ok {
		t.Fatal("bestMatch: expected match, got none")
	}
	if url != "https://example.com/leche.jpg" {
		t.Errorf("bestMatch URL = %q, want leche.jpg", url)
	}
}

func TestBestMatch_PartialMatch_BelowThreshold(t *testing.T) {
	index := ProductIndex{
		{Thumbnail: "https://example.com/aceite.jpg", Keywords: []string{"aceite", "oliva", "virgen", "extra"}},
	}
	// "aceite girasol" → only "aceite" matches (1/2 = 0.5 < minMatchScore 1.0)
	_, ok := bestMatch([]string{"aceite", "girasol"}, index)
	if ok {
		t.Error("bestMatch: expected no match for partial overlap, got one")
	}
}

func TestBestMatch_EmptyIndex(t *testing.T) {
	_, ok := bestMatch([]string{"leche"}, ProductIndex{})
	if ok {
		t.Error("bestMatch: expected no match on empty index, got one")
	}
}

func TestBestMatch_EmptyLocalKeywords(t *testing.T) {
	index := ProductIndex{
		{Thumbnail: "https://example.com/leche.jpg", Keywords: []string{"leche"}},
	}
	// Division by zero guard: len(localKW) == 0 → score always 0/0.
	// bestMatch should handle this gracefully (score = 0, no match).
	_, ok := bestMatch([]string{}, index)
	if ok {
		t.Error("bestMatch: expected no match for empty local keywords, got one")
	}
}

func TestBestMatch_PicksBestCandidate(t *testing.T) {
	index := ProductIndex{
		{Thumbnail: "https://example.com/yogur-natural.jpg", Keywords: []string{"yogur", "natural"}},
		{Thumbnail: "https://example.com/yogur-coco.jpg", Keywords: []string{"yogur", "coco"}},
	}
	// Local: "yogur natural" → [yogur, natural]; first entry scores 1.0, second 0.5.
	url, ok := bestMatch([]string{"yogur", "natural"}, index)
	if !ok {
		t.Fatal("bestMatch: expected match, got none")
	}
	if url != "https://example.com/yogur-natural.jpg" {
		t.Errorf("bestMatch picked wrong candidate: %q", url)
	}
}
