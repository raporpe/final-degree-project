package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	var ret [][]string

	client := http.Client{
		Timeout: 10,
	}

	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

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

	err = json.Unmarshal([]byte(body), &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
