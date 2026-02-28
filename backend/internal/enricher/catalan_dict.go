package enricher

import "strings"

// catalanToSpanish maps normalised Catalan tokens (as produced by normalise)
// to their Spanish equivalents.  The values are already in normalised form
// (lowercase, no accents) so they can be used directly in keyword matching.
//
// The dictionary covers the tokens found in Mercadona Catalunya receipts.
// Entries with an empty-string value indicate tokens that should be dropped
// entirely (i.e. they carry no discriminating information in Spanish).
var catalanToSpanish = map[string]string{
	// Dairy
	"llet":   "leche",
	"semi":   "semidesnatada",
	"llact":  "", // "s/lact" → sin lactosa; the token "lact" (split off) is also dropped
	"lact":   "", // residual from "s/lact" after normalise splits on "/"
	"ous":    "huevos",
	"clara":  "clara",
	"pasteu": "", // pasteuritzada – drop, not used in Spanish product names

	// Meat & poultry
	"pollastre": "pollo",
	"pit":       "pollo", // "pit" = pechuga/pit de pollastre
	"llom":      "lomo",
	"pernil":    "jamon",
	"burger":    "hamburguesa",
	"bovi":      "vacuno",
	"gruixuda":  "gruesa", // "burger gruixuda" = hamburguesa gruesa

	// Fish & seafood
	"tonyina":  "atun",
	"salmo":    "salmon",
	"fumat":    "ahumado",
	"verat":    "caballa",
	"filet":    "filetes",
	"musclo":   "mejillon",
	"escabetx": "escabeche",
	"paqu":     "", // "paquete" – drop

	// Vegetables
	"carbasso": "calabacin",
	"pebrot":   "pimiento",
	"ceba":     "cebolla",
	"espinaca": "espinaca",
	"esparrec": "esparrago",
	"xampinyo": "champinon",
	"brots":    "brotes",
	"tendres":  "tiernos",
	"llima":    "lima",
	"verd":     "verde",   // "carbassó verd" = calabacín verde
	"mitja":    "mediano", // "espàrrec mitjà" = espárrago mediano
	"mitjana":  "mediana",
	"patata":   "patata", // already Spanish; listed to aid completeness
	"tub":      "",       // "ceba tub" = cebolla en rama; drop modifier

	// Fruit
	"manz": "manzana",

	// Legumes & grains
	"cigro":   "garbanzo",
	"cuit":    "cocido",
	"nyoquis": "gnocchi",
	"arros":   "arroz",

	// Dairy / cheese
	"formatge":  "queso",
	"formatges": "queso",
	"provolone": "provolone",
	"mescla":    "mezcla",

	// Bread & bakery
	"xapata": "chapata",
	"torrat": "tostado",
	"panses": "pasas",
	"vidre":  "", // "pa de vidre" = chapata; drop redundant token

	// Oils & condiments
	"oliva":  "aceituna",
	"pinyol": "", // "sense pinyol" = sin hueso; drop

	// Drinks
	"cervesa": "cerveza",

	// Snacks & sweets
	"cacau":   "cacao",
	"patates": "patatas",

	// Eggs specifics
	"pages": "campo", // "ous pagès" = huevos de campo
	"pags":  "campo",

	// Misc
	"hummus":      "hummus",
	"truita":      "tortilla",
	"crema":       "crema",
	"curri":       "curry",
	"light":       "light",
	"classic":     "clasico",
	"fam":         "", // abbreviation with no clear translation; drop
	"nat":         "", // abbreviation; drop
	"granel":      "", // "a granel" = bulk; drop
	"talls":       "", // "talls" = slices; already implied by "filetes"; drop
	"engreix":     "", // "d'engreixament" = grain-fed; drop
	"piquillo":    "piquillo",
	"trico":       "tricolor",
	"ultra":       "ultra",
	"white":       "white",
	"net":         "", // "net" = clean/washed; drop
	"laminat":     "laminado",
	"baby":        "baby",
	"extrafinatr": "extrafino", // truncated "extrafina Trébol"
	"trev":        "",          // brand suffix; drop
	"sense":       "",          // "sense" = sin; drop (e.g. "sense pinyol")
}

// translateCatalan replaces Catalan tokens in a normalised product name with
// their Spanish equivalents.  Tokens not present in the dictionary are kept
// unchanged.  Empty-string values cause the token to be dropped.
//
// The input must already be normalised (output of normalise).
func translateCatalan(normalised string) string {
	tokens := strings.Fields(normalised)
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if replacement, found := catalanToSpanish[tok]; found {
			if replacement != "" {
				out = append(out, replacement)
			}
			// empty replacement → drop token
		} else {
			out = append(out, tok)
		}
	}
	return strings.Join(out, " ")
}
