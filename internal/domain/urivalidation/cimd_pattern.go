package urivalidation

import (
	"fmt"
	"net/url"
	"strings"
)

const cimdClientURIPatternPlaceholder = "wildcard"

func ValidateCIMDClientURIRegistration(raw string) error {
	if !strings.Contains(raw, "*") {
		return ValidateCIMDClientURL(raw)
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid CIMD client URI pattern: %w", err)
	}

	segments := strings.Split(u.EscapedPath(), "/")
	wildcardCount := 0
	for _, segment := range segments {
		if segment == "*" {
			wildcardCount++
		}
	}
	if wildcardCount != strings.Count(raw, "*") {
		return fmt.Errorf("CIMD client URI wildcard must replace a complete path segment")
	}

	return ValidateCIMDClientURL(strings.ReplaceAll(raw, "*", cimdClientURIPatternPlaceholder))
}

func MatchesCIMDClientURI(pattern, candidate string) bool {
	if err := ValidateCIMDClientURIRegistration(pattern); err != nil || !strings.Contains(pattern, "*") {
		return false
	}
	if err := ValidateCIMDClientURL(candidate); err != nil {
		return false
	}

	patternURL, err := url.Parse(pattern)
	if err != nil {
		return false
	}
	candidateURL, err := url.Parse(candidate)
	if err != nil {
		return false
	}
	if patternURL.Scheme != candidateURL.Scheme || patternURL.Host != candidateURL.Host || patternURL.RawQuery != candidateURL.RawQuery {
		return false
	}

	patternSegments := strings.Split(patternURL.EscapedPath(), "/")
	candidateSegments := strings.Split(candidateURL.EscapedPath(), "/")
	if len(patternSegments) != len(candidateSegments) {
		return false
	}
	for i, patternSegment := range patternSegments {
		if patternSegment == "*" {
			if candidateSegments[i] == "" || candidateSegments[i] == "*" {
				return false
			}
			continue
		}
		if patternSegment != candidateSegments[i] {
			return false
		}
	}

	return true
}
