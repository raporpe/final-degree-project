package main

import (
	"reflect"
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

func (s Record) IsActive() bool {
	// Mask search for XOOX XOXO OXOX
	for i := 0; i < len(s.Record)-3; i++ {
		toCompare := []bool{s.Record[i], s.Record[i+1], s.Record[i+2], s.Record[i+3]}

		a := []bool{true, false, false, true}
		b := []bool{true, false, true, false}
		c := []bool{false, true, false, true}

		aC := reflect.DeepEqual(toCompare, a)
		bC := reflect.DeepEqual(toCompare, b)
		cC := reflect.DeepEqual(toCompare, c)

		if aC {
			s.Record[i] = true
			s.Record[i+1] = true
			s.Record[i+2] = true
			s.Record[i+3] = true
		}

		if bC {
			s.Record[i] = true
			s.Record[i+1] = true
			s.Record[i+2] = true
			s.Record[i+3] = false
		}

		if cC {
			s.Record[i] = false
			s.Record[i+1] = true
			s.Record[i+2] = true
			s.Record[i+3] = true
		}

	}

	// Determine if active
	rowActive := 0
	for i := len(s.Record) - 1; i > 0; i-- {
		if s.Record[i] {
			rowActive++
		} else {
			break
		}
	}

	//Count total active
	active := 0
	for _, s := range s.Record {
		if s {
			active++
		}
	}

	rowActiveRatio := float64(rowActive) / float64(len(s.Record))
	activeRatio := float64(active) / float64(len(s.Record))

	if activeRatio > 0.65 {
		return true
	}

	if rowActiveRatio > 0.3 {
		return true
	}

	return false
}

func (s Record) applyMask(mask []bool) []bool {
	// Mask must be at least 3
	if len(mask) < 3 {
		return s.Record
	}

	for i := 0; i < len(s.Record)-len(mask)+1; i++ {
		match := true
		for idx, _ := range mask {
			if s.Record[idx] != mask[idx] {
				match = false
			}
		}

		// Set to true if there was a mask match
		if match {
			for idx, _ := range mask {
				s.Record[idx] = true
			}
		}

	}
	return s.Record
}
