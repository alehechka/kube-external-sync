package common

import (
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
)

type Replicator interface {
	Run()
	Synced() bool
	NamespaceAdded(ns *v1.Namespace)
}

func MatchStrictRegex(pattern, str string) (bool, error) {
	return regexp.MatchString(BuildStrictRegex(pattern), str)
}

func BuildStrictRegex(regex string) string {
	reg := strings.TrimSpace(regex)
	if !strings.HasPrefix(reg, "^") {
		reg = "^" + reg
	}
	if !strings.HasSuffix(reg, "$") {
		reg = reg + "$"
	}
	return reg
}
