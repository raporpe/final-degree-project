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

func TestCluster(t *testing.T) {

	m1 := []MacDigest{
		{
			Mac: "aa:3c:1a:e3:5a:50", // Cluster A
			Tags: []int{
				10,
				50,
				5,
				1,
			},
		},
		{
			Mac: "ab:3c:1a:e3:5a:50", // Cluster A
			Tags: []int{
				10,
				50,
				5,
				1,
			},
		},
		{
			Mac: "zz:3c:1a:e3:5a:50", // Similar to ab:3c:1a:e3:5a:50 but different ordering
			Tags: []int{ // (test not in cluster A)
				10,
				5,
				50,
				1,
			},
		},
		{
			Mac: "bb:01:1a:e3:5a:50", // Cluster B
			Tags: []int{
				10,
				50,
				1,
				2,
			},
		},
		{
			Mac: "bb:02:1a:e3:5a:50", // Cluster B
			Tags: []int{
				10,
				50,
				1,
				2,
			},
		},
		{
			Mac: "bb:03:1a:e3:5a:50", // Cluster B
			Tags: []int{
				10,
				50,
				1,
				2,
			},
		},
		{
			Mac: "cc:01:1a:e3:5a:50", // Unique but similar
			Tags: []int{
				1,
				2,
				3,
				4,
			},
		},
		{
			Mac: "dd:01:1a:e3:5a:50", // Unique but similar
			Tags: []int{
				1,
				2,
				4,
				3,
			},
		},
		{
			Mac: "ee:01:1a:e3:5a:50", // Unique but similar
			Tags: []int{
				5,
				2,
				4,
				3,
			},
		},
	}

	want := [][]string{
		{
			"aa:3c:1a:e3:5a:50",
			"ab:3c:1a:e3:5a:50",
		},
		{
			"zz:3c:1a:e3:5a:50",
		},
		{
			"bb:01:1a:e3:5a:50",
			"bb:02:1a:e3:5a:50",
			"bb:03:1a:e3:5a:50",
		},
		{
			"cc:01:1a:e3:5a:50",
		},
		{
			"dd:01:1a:e3:5a:50",
		},
		{
			"ee:01:1a:e3:5a:50",
		},
	}

	have := ClusteringVendorTags(m1)

	if !cmp.Equal(want, have) {
		t.Fatalf("Error in Clustering, \ngot %v, \nwanted %v", have, want)
	}

}
