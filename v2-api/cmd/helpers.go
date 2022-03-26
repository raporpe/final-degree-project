package main

func howManyTrue(slice []bool) int {
	count := 0
	for _, b := range slice {
		if b {
			count++
		}
	}
	return count
}
