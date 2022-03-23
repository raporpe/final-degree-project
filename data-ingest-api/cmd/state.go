package main

import (
	"strings"
)

type Record struct {
	Record []bool `json:"record"`
}

func NewRecord(s string) *Record {
	newState := Record{
		Record: make([]bool, len(s)),
	}

	for index, char := range s {
		newState.Record[index] = string(char) == "1"
	}

	return &newState
}

func (s Record) String() string {
	var sb strings.Builder
	for _, bool := range s.Record {
		if bool {
			sb.WriteString("1")
		} else {
			sb.WriteString("0")
		}
	}

	return sb.String()
}

func (s *Record) Set(index int, value bool) {
	s.Record[index] = value
}

func (s Record) IsEmpty() bool {
	for _, bool := range s.Record {
		if bool {
			// Not empty
			return false
		}
	}
	return true
}

func (s *Record) Union(otherState *Record) {
	for index, os := range otherState.Record {
		s.Record[index] = s.Record[index] || os
	}
}

func (s *Record) Shift(shift int) {
	// Left shift in the state
	for i := 0; i < shift; i++ {
		// Add false at the end
		s.Record = append(s.Record, false)

		// Delete the first one
		s.Record = s.Record[1:len(s.Record)]
	}
}
