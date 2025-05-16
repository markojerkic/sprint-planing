package services

import (
	"fmt"
	"testing"

	"github.com/markojerkic/spring-planing/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestRouteMatches(t *testing.T) {
	testCases := []struct {
		routeA  service.Route
		routeB  service.Route
		matches bool
	}{
		{
			routeA:  service.Route("a/b"),
			routeB:  service.Route("a/b"),
			matches: true,
		},
		{
			routeA:  service.Route("a/*"),
			routeB:  service.Route("a/b"),
			matches: true,
		},
		{
			routeA:  service.Route("a/b"),
			routeB:  service.Route("a/*"),
			matches: true,
		},
		{
			routeA:  service.Route("a/*"),
			routeB:  service.Route("a/*"),
			matches: true,
		},
		{
			routeA:  service.Route("a/*/b"),
			routeB:  service.Route("a/*/a"),
			matches: false,
		},
		{
			routeA:  service.Route("a"),
			routeB:  service.Route("a/*/a"),
			matches: false,
		},
		{
			routeA:  service.Route("a"),
			routeB:  service.Route("b"),
			matches: false,
		},
		{
			routeA:  service.Route("a/*"),
			routeB:  service.Route("a/b"),
			matches: true,
		},
		{
			routeA:  service.Route("a/b"),
			routeB:  service.Route("a/*"),
			matches: true,
		},
		{
			routeA:  service.Route("*/b"),
			routeB:  service.Route("a/*"),
			matches: true,
		},
		{
			routeA:  service.Route("*/*"),
			routeB:  service.Route("a/*"),
			matches: true,
		},
	}

	for _, testCase := range testCases {
		a := testCase.routeA
		b := testCase.routeB
		matches := a.Matches(b)

		assert.Equal(t, testCase.matches, matches, fmt.Sprintf("%s should match %s: %t, but received %t", string(a), string(b), testCase.matches, matches))
	}
}
