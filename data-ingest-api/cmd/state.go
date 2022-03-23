package main

import (
	"strings"
)

type State struct {
	state []bool
}

func NewState(s string) *State {
	newState := State{
		state: make([]bool, len(s)),
	}

	for index, char := range s {
		newState.state[index] = string(char) == "1"
	}

	return &newState
}

func (s State) String() string {
	var sb strings.Builder
	for _, bool := range s.state {
		if bool {
			sb.WriteString("1")
		} else {
			sb.WriteString("0")
		}
	}

	return sb.String()
}

func (s *State) Set(index int, value bool) {
	s.state[index] = value
}

func (s State) IsEmpty() bool {
	for _, bool := range s.state {
		if bool {
			// Not empty
			return false
		}
	}
	return true
}

func (s *State) Union(otherState *State) {
	for index, os := range otherState.state {
		s.state[index] = s.state[index] || os
	}
}

func (s *State) Shift(shift int) {
	// Left shift in the state
	for i := 0; i < shift; i++ {
		// Add false at the end
		s.state = append(s.state, false)

		// Delete the first one
		s.state = s.state[1:len(s.state)]
	}
}
