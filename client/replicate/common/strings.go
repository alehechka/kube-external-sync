package common

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MustGetKey creates a key from Kubernetes resource in the format <namespace>/<name>
func MustGetKey(obj interface{}) string {
	if obj == nil {
		return ""
	}

	o := MustGetObject(obj)
	return fmt.Sprintf("%s/%s", o.GetNamespace(), o.GetName())

}

// MustGetObject casts the object into a Kubernetes `metav1.Object`
func MustGetObject(obj interface{}) metav1.Object {
	if obj == nil {
		return nil
	}

	if oma, ok := obj.(metav1.ObjectMetaAccessor); ok {
		return oma.GetObjectMeta()
	} else if o, ok := obj.(metav1.Object); ok {
		return o
	}

	panic(fmt.Errorf("Unknown type: %v", reflect.TypeOf(obj)))
}

func StringToPatternList(list string) (result []*regexp.Regexp) {
	for _, s := range strings.Split(list, ",") {
		r, err := CompileStrictRegex(s)
		if err != nil {
			log.WithError(err).Errorf("Invalid regex '%s' in namespace string %s: %v", s, list, err)
		} else {
			result = append(result, r)
		}
	}

	return
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

func MatchStrictRegex(pattern, str string) (bool, error) {
	return regexp.MatchString(BuildStrictRegex(pattern), str)
}

func CompileStrictRegex(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(BuildStrictRegex(pattern))
}

func PrepareTLD(namespace, tld string) string {
	subdomains := strings.Split(tld, ".")
	subdomains[0] = namespace

	return strings.Join(subdomains, ".")
}

func PrepareRouteMatch(namespace, match string) string {
	if !strings.Contains(match, "Host") {
		return match
	}

	parts := strings.Split(match, "`")

	isHost := false
	for index, part := range parts {
		if isHost {
			parts[index] = PrepareTLD(namespace, part)
			isHost = false
			continue
		}
		if index%2 == 0 && strings.Contains(part, "Host") {
			isHost = true
		}
	}

	return strings.Join(parts, "`")
}
