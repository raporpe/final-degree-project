package main

import (
	"time"
)

func howManyTrue(slice []bool) int {
	count := 0
	for _, b := range slice {
		if b {
			count++
		}
	}
	return count
}

func timePlusWindow(startTime time.Time, windows int, windowSize int) time.Time {
	return startTime.Add(time.Second * time.Duration(windowSize*windows))
}

func DeduplicateSlice[K comparable](slice []K) []K {
	var ret []K

	m := make(map[K]bool)
	for _, v := range slice {
		m[v] = true
	}

	for k := range m {
		ret = append(ret, k)
	}

	return ret
}
