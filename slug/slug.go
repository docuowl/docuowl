package slug

import "unicode"

func Slugify(s string) string {
	cutDash := true
	emitDash := false

	slug := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsNumber(r) || unicode.IsLetter(r) {
			if emitDash && !cutDash {
				slug = append(slug, '-')
			}
			slug = append(slug, unicode.ToLower(r))

			emitDash = false
			cutDash = false
			continue
		}
		switch r {
		case '/', '=':
			if len(slug) == 0 || slug[len(slug)-1] != r {
				slug = append(slug, r)
			}
			emitDash = false
			cutDash = true
		case '-', ',', '.', ' ', '_':
			emitDash = true
		default:
			if name, exists := runename[r]; exists {
				if !cutDash {
					slug = append(slug, '-')
				}
				slug = append(slug, []rune(name)...)
				cutDash = false
			}
			emitDash = true
		}
	}

	if len(slug) == 0 {
		return "-"
	}

	return string(slug)
}
