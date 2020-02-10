package utils

import "sort"

func IntersectInt(a []int, b []int) []int {
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

func UnionInt(a []int, b []int) []int {
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

func DifferenceInt(a []int, b []int) []int {
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

func IntersectString(a []string, b []string) []string {
	hash := make(map[string]bool)
	res := make([]string, 0)

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

func UnionString(a []string, b []string) []string {
	hash := make(map[string]bool)
	res := make([]string, 0)

	for _, i := range a {
		hash[i] = true
	}
	for _, i := range b {
		hash[i] = true
	}

	for key := range hash {
		res = append(res, key)
	}
	sort.Strings(res)
	return res
}

func DifferenceString(a []string, b []string) []string {
	hash := make(map[string]bool)
	res := make([]string, 0)

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
