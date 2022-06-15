package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/exp/constraints"
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

// Performs the clustering of the mac addresses
// It only takes into account the tags
func Clustering(m []MacDigest) [][]string {

	// Index the MacDigests by their tags
	macIndexByTags := make(map[string][]MacDigest)

	for _, v := range m {
		// Transform tags into string
		tag := SliceToString(v.Tags)

		// Index the MacDigest struct by the tag
		stored := macIndexByTags[tag]

		// Add the MacDigest struct
		if stored == nil {
			macIndexByTags[tag] = []MacDigest{v}
		} else {
			macIndexByTags[tag] = append(macIndexByTags[tag], v)
		}
	}

	// Now traverse the map and analyse those clusters
	// that have more than one value
	for _, cluster := range macIndexByTags {
		// If the cluster has more than one mac address
		if len(cluster) > 1 {
			fmt.Printf("Found a cluster with %v mac addresses! ", len(cluster))
			for _, macDigest := range cluster {
				fmt.Printf("%v,", macDigest)
			}
			fmt.Println()
		}
	}

	// Transform the map into the final returned type
	// From MacDigest to clusters of strings
	// These strings are mac addresses
	var ret [][]string
	for _, cluster := range macIndexByTags {

		inner := make([]string, 0)
		for _, vv := range cluster {
			inner = append(inner, vv.Mac)
		}

		ret = append(ret, inner)

	}

	return ret
}

func SliceToString[K int | float64 | int64](slice []K) string {
	ret := ""

	for _, v := range slice {
		ret += fmt.Sprintf("%v,", v)
	}

	return ret
}

func IsDeviceActive(presenceRecord []bool) bool {
	totalCounter := 0
	recentCounter := 0
	for i, v := range presenceRecord {
		if v {
			totalCounter += 1
			if float64(i) >= float64(len(presenceRecord))-float64(len(presenceRecord))/5.0 {
				recentCounter += 1
			}
		}
	}

	totalActivity := float64(totalCounter)/float64(len(presenceRecord)) > 0.6
	recentActivity := float64(recentCounter)/float64(len(presenceRecord)/5) > 0.6

	return totalActivity || (recentActivity && false)
}

// Merge cluster c2 into cluster c1
func ClusterMerge(c1 [][]string, c2 [][]string, shareThreshold float64) [][]string {

	// Index the cluster c1 for faster search
	index := make(map[string]int)
	for c1ClusterID, macs := range c1 {
		for _, mac := range macs {
			index[mac] = c1ClusterID
		}
	}

	// Traverse the second cluster and search for matches
	for c2ClusterID, c2Cluster := range c2 {
		mergeFlag := false
		for _, mac := range c2Cluster {
			// If the mac in c2 is present on c1
			// and the current c2 cluster has not already been merged
			c1ClusterID, exists := index[mac]
			if exists && !mergeFlag {
				// Analyze merge cluster and compare with current one (in c2)
				cluster1 := c1[c1ClusterID]
				cluster2 := c2[c2ClusterID]
				duplicates := GetNumberOfDuplicates(cluster1, cluster2)
				// Check if duplicates superate threshold
				if float64(duplicates)/float64(min(len(cluster1), len(cluster2))) > shareThreshold {
					// The cluster c2 is merged into c1
					c1[c1ClusterID] = DeduplicateSlice(append(c1[c1ClusterID], c2[c2ClusterID]...))
					mergeFlag = true
				}
			}
		}
		// If for this cluster in c2 there was not any merge,
		// append it to merge
		if !mergeFlag {
			c1 = append(c1, c2Cluster)
		}
	}

	return c1
}

func min[T constraints.Ordered](a T, b T) T {
	if a >= b {
		return b
	} else {
		return a
	}
}

func GetNumberOfDuplicates[K comparable](s1 []K, s2 []K) int {
	// Index s1
	s1Index := make(map[K]bool, 0)
	for _, v := range s1 {
		s1Index[v] = true
	}

	// Count the number of matches
	matches := 0

	for _, v := range s2 {
		// If s2 exists in s1, then mark
		if _, exists := s1Index[v]; exists {
			matches += 1
		}
	}

	return matches

}

func Optics(m []MacDigest) ([][]string, error) {

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	fmt.Printf("j: %v\n", string(j))

	req, err := http.NewRequest("POST", "http://optics/", bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("Not 200 response")
	}

	// Read the body and decode the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var opticsResult []int
	err = json.Unmarshal([]byte(body), &opticsResult)
	if err != nil {
		return nil, err
	}

	fmt.Printf("opticsResult: %v\n", opticsResult)

	ret := make([][]string, len(DeduplicateSlice(opticsResult)))

	for k, v := range opticsResult {
		if v < 0 {
			continue
		}
		ret[v] = append(ret[v], m[k].Mac)
	}

	return ret, nil
}

// Function that calls the clustering API at http://clustering:8000/cluster
// and returns the clusters in a list of lists of strings
func Clustering2(m []MacDigest) ([][]string, error) {

	if m == nil {
		return nil, nil
	}

	// If the lenght of MacDigest is less than 5, convert to return type
	if len(m) < 5 {
		ret := make([][]string, 0)
		for _, v := range m {
			ret = append(ret, []string{v.Mac})
		}
		return ret, nil
	}

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "http://clustering:8000/cluster", bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Error calling clustering API: " + err.Error())
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("Not 200 response when calling clustering API")
	}

	// Read the body and decode the response into a list of lists of strings
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error decoding the response from clustering API: " + err.Error())
	}

	var clusters [][]string
	err = json.Unmarshal([]byte(body), &clusters)
	if err != nil {
		return nil, errors.New("Error decoding the response from clustering API: " + err.Error())
	}

	return clusters, nil

}
