package admin

import "strings"

func requestsCanonicalReferences(prefer string) bool {
	for _, preference := range strings.Split(prefer, ",") {
		if strings.TrimSpace(preference) == "reference-id=canonical" {
			return true
		}
	}
	return false
}
