package utils

import (
	"regexp"
)

func GetAllMatchesByNames(reg *regexp.Regexp, s string) map[string]string {
	n1 := reg.SubexpNames()
	r2 := reg.FindAllStringSubmatch(s, -1)[0]

	md := map[string]string{}
	for i, n := range r2 {
		md[n1[i]] = n
	}

	return md
}
