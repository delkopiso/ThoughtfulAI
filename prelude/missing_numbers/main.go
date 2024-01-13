package main

import (
	"fmt"
	"sort"
)

func main() {
	fmt.Println("{2,3,4,4,2,5,2,5}:", findMissing([]int{2, 3, 4, 4, 2, 5, 2, 5}))
	fmt.Println("{3,4,4,5,5,1,-1,0}:", findMissing([]int{3, 4, 4, 5, 5, 1, -1, 0}))
}

// Given an unordered array of integers, find the first missing positive number
func findMissing(arr []int) int {
	sort.Ints(arr)
	minimum := 1
	for _, element := range arr {
		if element == minimum {
			minimum += 1
		}
		if element > minimum {
			return minimum
		}
	}
	return 0
}
