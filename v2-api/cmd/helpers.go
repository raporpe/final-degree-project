package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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
