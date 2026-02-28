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

// Dice = 2·|A∩B| / (|A|+|B|).  minMatchScore = 0.5.

func TestBestMatch_ExactMatch(t *testing.T) {
	index := ProductIndex{
		{Thumbnail: "https://example.com/leche.jpg", Keywords: []string{"leche", "entera"}},
		{Thumbnail: "https://example.com/pan.jpg", Keywords: []string{"pan", "integral", "molde"}},
	}
	// local=[leche,entera], entry=[leche,entera] → matched=2, Dice=2·2/(2+2)=1.0 ≥ 0.5 ✓
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
		// entry has 4 keywords; local has 2; only 1 shared → Dice = 2·1/(2+4) = 0.33 < 0.5
		{Thumbnail: "https://example.com/aceite.jpg", Keywords: []string{"aceite", "oliva", "virgen", "extra"}},
	}
	_, ok := bestMatch([]string{"aceite", "girasol"}, index)
	if ok {
		t.Error("bestMatch: expected no match for low Dice score, got one")
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
	_, ok := bestMatch([]string{}, index)
	if ok {
		t.Error("bestMatch: expected no match for empty local keywords, got one")
	}
}

func TestBestMatch_PicksBestCandidate(t *testing.T) {
	index := ProductIndex{
		// local=[yogur,natural], entry1=[yogur,natural] → Dice=2·2/(2+2)=1.0
		{Thumbnail: "https://example.com/yogur-natural.jpg", Keywords: []string{"yogur", "natural"}},
		// local=[yogur,natural], entry2=[yogur,coco] → matched=1, Dice=2·1/(2+2)=0.5
		{Thumbnail: "https://example.com/yogur-coco.jpg", Keywords: []string{"yogur", "coco"}},
	}
	url, ok := bestMatch([]string{"yogur", "natural"}, index)
	if !ok {
		t.Fatal("bestMatch: expected match, got none")
	}
	if url != "https://example.com/yogur-natural.jpg" {
		t.Errorf("bestMatch picked wrong candidate: %q", url)
	}
}

// TestBestMatch_DiceRejectsFalsePositive verifies that the Dice metric prevents
// a single shared keyword from matching a catalogue entry with many more keywords.
// Concretely: local=["patata"] vs entry=["patatas","fritas","onduladas","pringles"]
// Dice = 2·1/(1+4) = 0.4 < 0.5 → no match.
func TestBestMatch_DiceRejectsFalsePositive(t *testing.T) {
	index := ProductIndex{
		{Thumbnail: "https://example.com/pringles.jpg", Keywords: []string{"patatas", "fritas", "onduladas", "pringles"}},
	}
	_, ok := bestMatch([]string{"patata"}, index)
	if ok {
		t.Error("bestMatch: Dice should reject 'patata' matching Pringles (4 extra keywords)")
	}
}

// TestBestMatch_DiceAcceptsCloseMatch verifies a local product with few keywords
// matches a catalogue entry with similar keywords.
// local=["patata"] vs entry=["patata","hacendado"] → Dice=2·1/(1+2)=0.67 ≥ 0.5 ✓
func TestBestMatch_DiceAcceptsCloseMatch(t *testing.T) {
	index := ProductIndex{
		{Thumbnail: "https://example.com/patata.jpg", Keywords: []string{"patata", "hacendado"}},
		{Thumbnail: "https://example.com/pringles.jpg", Keywords: []string{"patatas", "fritas", "onduladas", "pringles"}},
	}
	url, ok := bestMatch([]string{"patata"}, index)
	if !ok {
		t.Fatal("bestMatch: expected match for patata vs patata-hacendado, got none")
	}
	if url != "https://example.com/patata.jpg" {
		t.Errorf("bestMatch picked wrong candidate: %q", url)
	}
}

// ---------- translateCatalan ----------

func TestTranslateCatalan(t *testing.T) {
	tests := []struct {
		name  string
		input string // already normalised (output of normalise)
		want  string
	}{
		{
			name:  "leche semidesnatada sin lactosa",
			input: "llet semi s llact",
			// "s" is not in the dictionary (it's a stop word, filtered later by keywords())
			// "llact" maps to "" so it is dropped; "s" is passed through unchanged
			want: "leche semidesnatada s",
		},
		{
			name:  "huevos de campo",
			input: "12 ous pages",
			want:  "12 huevos campo",
		},
		{
			name:  "queso rallado 4 quesos",
			input: "mescla 4 formatges",
			want:  "mezcla 4 queso",
		},
		{
			name:  "atun claro",
			input: "tonyina clara natura",
			want:  "atun clara natura",
		},
		{
			name:  "champinon laminado",
			input: "xampinyo net laminat",
			want:  "champinon laminado",
		},
		{
			name:  "salmon ahumado",
			input: "salmo fumat",
			want:  "salmon ahumado",
		},
		{
			name:  "tortilla patata cebolla",
			input: "truita patata ceba",
			want:  "tortilla patata cebolla",
		},
		{
			name:  "chapata",
			input: "xapata vidre",
			want:  "chapata",
		},
		{
			name:  "non-catalan tokens are preserved",
			input: "coca cola zero",
			want:  "coca cola zero",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		// ── New dictionary entries ────────────────────────────────────────────
		{
			name:  "pollo pechuga",
			input: "pollastre pit",
			want:  "pollo pechuga",
		},
		{
			name:  "pimiento rojo",
			input: "pebrot vermell",
			want:  "pimiento rojo",
		},
		{
			name:  "lavavajillas detergente",
			input: "rentaplats detergent",
			want:  "lavavajillas detergente",
		},
		{
			name:  "limon sin hueso",
			input: "llimona sense pinyol",
			// "sense" and "pinyol" map to "" → dropped
			want: "limon",
		},
		{
			name:  "cerveza lata",
			input: "cervesa llauna",
			want:  "cerveza lata",
		},
		{
			name:  "yogur griego",
			input: "iogurt grec",
			want:  "yogur griego",
		},
		{
			name:  "helado mini",
			input: "gelat mini",
			want:  "helado mini",
		},
		{
			name:  "aceite oliva virgen extra",
			input: "oli oliva verge extra",
			want:  "aceite aceituna virgen extra",
		},
		{
			name:  "champu argan",
			input: "xampu argan",
			want:  "champu argan",
		},
		{
			name:  "detergente suavizante",
			input: "detergent suavitzant",
			want:  "detergente suavizante",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateCatalan(tt.input)
			if got != tt.want {
				t.Errorf("translateCatalan(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
