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

func DeduplicateSlice(slice []string) []string {
	var ret []string

	m := make(map[string]string)
	for _, v := range slice {
		m[v] = ""
	}

	for k := range m {
		ret = append(ret, k)
	}

	return ret
}
