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
			"AAB", // Match result
			"AAD", // Match result
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

// Test the IsDeviceActive function
// This function is responsible for detecting which devices
// are active based on a presence boolean slice
// It is crucual for filtering transient devices
func TestIsDeviceActive(t *testing.T) {

	// Positive in total activity
	// Negative in recent activity
	t1 := []bool{
		true,
		false,
		true,
		true,
		false,

		true,
		true,
		true,
		false,
		false,

		false,
		false,
		false,
		true,
		false,
	}

	have := IsDeviceActive(t1)
	want := true

	if want != have {
		t.Fatalf("IsDeviceActive error: got: %v, wanted: %v ", have, want)
	}

	// Negative in both
	t2 := []bool{
		false,
		false,
		true,
		false,
		false,

		false,
		false,
		false,
		false,
		false,

		false,
		true,
		false,
		true,
		false,
	}

	have = IsDeviceActive(t2)
	want = false

	if want != have {
		t.Fatalf("IsDeviceActive error: got: %v, wanted: %v ", have, want)
	}

	// Positive in recent
	t3 := []bool{
		false,
		false,
		false,
		false,
		false,

		false,
		false,
		false,
		false,
		false,

		true,
		true,
		true,
		true,
		false,
	}

	have = IsDeviceActive(t3)
	want = true

	if want != have {
		t.Fatalf("IsDeviceActive error: got: %v, wanted: %v ", have, want)
	}

}
