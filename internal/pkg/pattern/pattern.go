package pattern

import (
	"regexp"
)

func GetNamedStringFromText(s string, regexpPattern string, name string) string {
	r, err := regexp.Compile(regexpPattern)
	if err != nil {
		panic(err.Error())
	}

	match := r.FindStringSubmatch(s)

	for i, k := range r.SubexpNames() {
		if i != 0 && k != "" {
			if k == name {
				if i < len(match) {
					return match[i]
				}
			}
		}
	}

	return ""
}
