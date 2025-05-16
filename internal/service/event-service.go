package service

import "strings"

type Route string

func (r Route) Matches(route Route) bool {
	requestSegments := strings.Split(string(route), "/")
	mySegments := strings.Split(string(r), "/")

	if len(requestSegments) != len(mySegments) {
		return false
	}

	i := 0
	j := 0

	for i < len(requestSegments) && j < len(mySegments) {
		requestSegment := requestSegments[i]
		mySegment := mySegments[j]

		if requestSegment == "*" || mySegment == "*" {
			i += 1
			j += 1
		} else if requestSegment == mySegment {
			i += 1
			j += 1
		} else {
			return false
		}
	}

	return true
}
