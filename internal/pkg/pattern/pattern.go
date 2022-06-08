package pattern

import (
	"regexp"
	"strings"
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

func GetNamedMapFromText(s string, regexpPattern string, names []string) map[string]string {
	r, err := regexp.Compile(regexpPattern)
	if err != nil {
		panic(err.Error())
	}

	result := make(map[string]string)
	for _, v := range names {
		result[v] = ""
	}

	match := r.FindStringSubmatch(s)

	for i, k := range r.SubexpNames() {
		if i != 0 && k != "" {
			if _, ok := result[k]; ok {
				if i < len(match) {
					result[k] = match[i]
				}
			}
		}
	}

	return result
}

func IsHttpUrl(s string) bool {
	ls := strings.ToLower(s)
	return strings.HasPrefix(ls, "https://") || strings.HasPrefix(ls, "http://")
}

func IsGithubUrl(s string) bool {
	ls := strings.ToLower(s)
	return strings.HasPrefix(ls, "https://github.com") || strings.HasPrefix(ls, "http://github.com")
}
