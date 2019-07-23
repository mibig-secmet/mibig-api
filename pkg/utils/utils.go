package utils

import "sort"

func Intersect(a []int, b []int) []int {
	hash := make(map[int]bool)
	res := make([]int, 0)

	for _, i := range a {
		hash[i] = true
	}

	for _, i := range b {
		if _, ok := hash[i]; ok {
			res = append(res, i)
		}
	}
	return res
}

func Union(a []int, b []int) []int {
	hash := make(map[int]bool)
	res := make([]int, 0)

	for _, i := range a {
		hash[i] = true
	}
	for _, i := range b {
		hash[i] = true
	}

	for key := range hash {
		res = append(res, key)
	}
	sort.Ints(res)
	return res
}

func Difference(a []int, b []int) []int {
	hash := make(map[int]bool)
	res := make([]int, 0)

	for _, i := range b {
		hash[i] = true
	}

	for _, i := range a {
		if _, ok := hash[i]; !ok {
			res = append(res, i)
		}
	}
	return res
}
