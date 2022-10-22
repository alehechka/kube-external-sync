package traefik

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PrepareRouteMatch_NoHost(t *testing.T) {
	match := "PathPrefix(`/`)"
	newMatch := PrepareRouteMatch("default", match, "")
	assert.Equal(t, match, newMatch)
}

func Test_PrepareRouteMatch_OneHost(t *testing.T) {
	match := "Host(`subdomain.example.com`)"
	newMatch := PrepareRouteMatch("default", match, "")
	assert.Equal(t, "Host(`default.example.com`)", newMatch)
}

func Test_PrepareRouteMatch_OneHost_WithFallback(t *testing.T) {
	match := "Host(`subdomain.example.com`)"
	newMatch := PrepareRouteMatch("default", match, "*.other.com")
	assert.Equal(t, "Host(`default.other.com`)", newMatch)
}

func Test_PrepareRouteMatch_TwoHosts(t *testing.T) {
	match := "Host(`subdomain.example.com`) || Host(`subdomain.placeholder.com`)"
	newMatch := PrepareRouteMatch("default", match, "")
	assert.Equal(t, "Host(`default.example.com`) || Host(`default.placeholder.com`)", newMatch)
}

func Test_PrepareRouteMatch_HostPathPrefix(t *testing.T) {
	match := "Host(`subdomain.example.com`) && PathPrefix(`/`)"
	newMatch := PrepareRouteMatch("default", match, "")
	assert.Equal(t, "Host(`default.example.com`) && PathPrefix(`/`)", newMatch)
}

func Test_PrepareRouteMatch_HostPathPrefixHost(t *testing.T) {
	match := "(Host(`subdomain.example.com`) && PathPrefix(`/`)) || Host(`subdomain.placeholder.com`)"
	newMatch := PrepareRouteMatch("default", match, "")
	assert.Equal(t, "(Host(`default.example.com`) && PathPrefix(`/`)) || Host(`default.placeholder.com`)", newMatch)
}

func Test_PrepareRouteMatch_MultiHosts(t *testing.T) {
	match := "Host(`subdomain.example.com`, `other.example.com`)"
	newMatch := PrepareRouteMatch("default", match, "")
	assert.Equal(t, "Host(`default.example.com`, `default.example.com`)", newMatch)
}

func Test_PrepareRouteMatch_Complex(t *testing.T) {
	match := "(Host(`subdomain.example.com`, `other.example.com`) && PathPrefix(`/`)) || Host(`subdomain.placeholder.com`, `other.placeholder.com`)"
	newMatch := PrepareRouteMatch("default", match, "")
	assert.Equal(t, "(Host(`default.example.com`, `default.example.com`) && PathPrefix(`/`)) || Host(`default.placeholder.com`, `default.placeholder.com`)", newMatch)
}

func Test_PrepareRouteMatch_NewLineSuffix(t *testing.T) {
	match := "Host(`example.com`) && Path(`/path1`,`/path2`,`/path3`)\n"
	newMatch := PrepareRouteMatch("default", match, "*.example.com")
	assert.Equal(t, "Host(`default.example.com`) && Path(`/path1`,`/path2`,`/path3`)", newMatch)
}

func Test_prepareDomainStrings(t *testing.T) {
	assert.Equal(t, "(`default.example.com`", prepareDomainStrings("default", "(`subdomain.example.com`", ""))
	assert.Equal(t, "`default.example.com`", prepareDomainStrings("default", "`subdomain.example.com`", ""))
	assert.Equal(t, "`default.example.com`, `default.placeholder.com`", prepareDomainStrings("default", "`subdomain.example.com`, `subdomain.placeholder.com`", ""))
}

func Test_prepareDomainStrings_WithFallback(t *testing.T) {
	assert.Equal(t, "(`default.other.com`", prepareDomainStrings("default", "(`subdomain.example.com`", "*.other.com"))
	assert.Equal(t, "`default.other.com`", prepareDomainStrings("default", "`subdomain.example.com`", "*.other.com"))
	assert.Equal(t, "`default.other.com`, `default.other.com`", prepareDomainStrings("default", "`subdomain.example.com`, `subdomain.placeholder.com`", "*.other.com"))
}

func Test_uncutSplit(t *testing.T) {
	assert.Equal(t, []string{}, uncutSplit("", parenthesisDelimiter))

	assert.Equal(t, []string{"", "("}, uncutSplit("(", parenthesisDelimiter))
	assert.Equal(t, []string{"", ")"}, uncutSplit(")", parenthesisDelimiter))

	assert.Equal(t, []string{"", "(Host", ")"}, uncutSplit("(Host)", parenthesisDelimiter))
	assert.Equal(t, []string{"Host", "(`myapp.example.com`", ")"}, uncutSplit("Host(`myapp.example.com`)", parenthesisDelimiter))
	assert.Equal(t, []string{"Host", "(`myapp.example.com`", ") && PathPrefix", "(`/`", ")"}, uncutSplit("Host(`myapp.example.com`) && PathPrefix(`/`)", parenthesisDelimiter))
	assert.Equal(t,
		[]string{"", "(Host", "(`myapp.example.com`", ") && PathPrefix", "(`/`", ")", ") || Host", "(`other.example.com`", ")"},
		uncutSplit("(Host(`myapp.example.com`) && PathPrefix(`/`)) || Host(`other.example.com`)", parenthesisDelimiter))
}

func Test_parenthesisDelimiter(t *testing.T) {
	assert.True(t, parenthesisDelimiter('('))
	assert.True(t, parenthesisDelimiter(')'))
	assert.False(t, parenthesisDelimiter('-'))
}
