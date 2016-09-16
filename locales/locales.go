//go:generate go run mkgplocales.go

// Package locales provides list of GP locales based on
// https://github.com/GlotPress/gp-locales
package locales

func GetLocaleProp(locale string, prop string) string {
	if v, ok := locales[locale]; !ok {
		return ""
	} else {
		switch prop {
		case "EnglishName":
			return v.EnglishName
		case "NativeName":
			return v.NativeName
		case "LangCodeISO6391":
			return v.LangCodeISO6391
		case "LangCodeISO6392":
			return v.LangCodeISO6392
		case "LangCodeISO6393":
			return v.LangCodeISO6393
		default:
			return ""
		}
	}
}
