package traefik

import (
	"strings"

	"github.com/alehechka/kube-external-sync/client/replicate/common"
)

// PrepareRouteMatch replaces all domains with the namespaced version
func PrepareRouteMatch(namespace, match string, hostname string) string {
	if !strings.Contains(match, "Host") {
		return match
	}

	parenParts := uncutSplit(match, parenthesisDelimiter)

	isHost := false
	for index, part := range parenParts {
		if isHost {
			parenParts[index] = prepareDomainStrings(namespace, part, hostname)
			isHost = false
			continue
		}
		if strings.Contains(part, "Host") {
			isHost = true
		}
	}

	return strings.Join(parenParts, "")
}

func prepareDomainStrings(namespace, domainString, hostname string) string {
	parts := strings.Split(domainString, "`")

	isHost := false
	for index, part := range parts {
		if isHost {
			if len(hostname) > 0 {
				parts[index] = common.PrepareTLD(namespace, hostname)
			} else {
				parts[index] = common.PrepareTLD(namespace, part)
			}

			isHost = false
			continue
		}
		isHost = true
	}

	return strings.Join(parts, "`")
}

// uncutSplit splits a given string with the delimiter but keeps all characters intact.
// Use case is specifically for a multi-character delimiter
func uncutSplit(str string, delimiter func(rune) bool) (strs []string) {
	if len(str) == 0 {
		return make([]string, 0)
	}

	var prev int
	for index, char := range str {
		if delimiter(char) {
			strs = append(strs, str[prev:index])
			prev = index
		}
	}

	if delimiter(rune(str[len(str)-1])) {
		strs = append(strs, string(str[len(str)-1]))
	}

	return
}

// parenthesisDelimiter is a delimiter function for open and close parenthesis
func parenthesisDelimiter(r rune) bool {
	return r == '(' || r == ')'
}
