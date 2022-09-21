package common_test

import (
	"testing"

	"github.com/alehechka/kube-external-sync/client/replicate/common"
	"github.com/stretchr/testify/assert"
)

func Test_PrepareRouteMatch_NoHost(t *testing.T) {
	match := "PathPrefix(`/`)"
	newMatch := common.PrepareRouteMatch("default", match)
	assert.Equal(t, match, newMatch)
}

func Test_PrepareRouteMatch_OneHost(t *testing.T) {
	match := "Host(`subdomain.example.com`)"
	newMatch := common.PrepareRouteMatch("default", match)
	assert.Equal(t, "Host(`default.example.com`)", newMatch)
}

func Test_PrepareRouteMatch_TwoHosts(t *testing.T) {
	match := "Host(`subdomain.example.com`) || Host(`subdomain.placeholder.com`)"
	newMatch := common.PrepareRouteMatch("default", match)
	assert.Equal(t, "Host(`default.example.com`) || Host(`default.placeholder.com`)", newMatch)
}

func Test_PrepareRouteMatch_HostPathPrefix(t *testing.T) {
	match := "Host(`subdomain.example.com`) && PathPrefix(`/`)"
	newMatch := common.PrepareRouteMatch("default", match)
	assert.Equal(t, "Host(`default.example.com`) && PathPrefix(`/`)", newMatch)
}

func Test_PrepareRouteMatch_HostPathPrefixHost(t *testing.T) {
	match := "(Host(`subdomain.example.com`) && PathPrefix(`/`)) || Host(`subdomain.placeholder.com`)"
	newMatch := common.PrepareRouteMatch("default", match)
	assert.Equal(t, "(Host(`default.example.com`) && PathPrefix(`/`)) || Host(`default.placeholder.com`)", newMatch)
}
