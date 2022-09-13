package types

import (
	"regexp"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type Annotation struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	IsRegex bool   `json:"isRegex"`
}

func (annotation *Annotation) AnnotatesNamespace(namespace *v1.Namespace) bool {
	for key, value := range namespace.Annotations {
		if key == annotation.Key {
			if annotation.IsRegex {
				isMatch, err := regexp.MatchString(annotation.Value, value)
				if err != nil {
					log.Errorf("failed to compile pattern: %s", err.Error())
				}
				if isMatch {
					return true
				}
			} else if value == annotation.Value {
				return true
			}
		}
	}

	return false
}

type Annotations []Annotation

func (annotations Annotations) AnnotatesNamespace(namespace *v1.Namespace) bool {
	for _, annotation := range annotations {
		if annotation.AnnotatesNamespace(namespace) {
			return true
		}
	}

	return false
}

func (annotations Annotations) IsEmpty() bool {
	return len(annotations) == 0
}
