package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Test the ClusterMergeFunction
func TestClusterMerge(t *testing.T) {

	c1 := [][]string{
		{
			"AAB", // Match
			"AAD", // Match
			"AAC",
		},
		{
			"QA",
			"QB",
			"QC",
		},
	}

	c2 := [][]string{
		{
			"ZA",
			"ZB",
		},
		{
			"AAB", // Match
			"AAD", // Match
			"AAZ",
		},
	}

	have := ClusterMerge(c1, c2, 0.33)

	want := [][]string{
		{
			"AAB", // Match
			"AAD", // Match
			"AAC",
			"AAZ",
		},
		{
			"QA",
			"QB",
			"QC",
		},
		{
			"ZA",
			"ZB",
		},
	}

	if !cmp.Equal(want, have) {
		t.Fatalf("ClusterMerge error: got: %v, wanted: %v ", have, want)
	}
}

//// TestHelloEmpty calls greetings.Hello with an empty string,
//// checking for an error.
//func TestHelloEmpty(t *testing.T) {
//	msg, err := Hello("")
//	if msg != "" || err == nil {
//		t.Fatalf(`Hello("") = %q, %v, want "", error`, msg, err)
//	}
//}
