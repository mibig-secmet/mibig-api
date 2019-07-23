package utils

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestIntersect(t *testing.T) {
	var tests = []struct {
		a        []int
		b        []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 4, 6}, []int{2}},
	}

	for _, tt := range tests {
		actual := Intersect(tt.a, tt.b)
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("Intersect(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, actual)
		}
	}
}

func TestUnion(t *testing.T) {
	var tests = []struct {
		a        []int
		b        []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 4, 6}, []int{1, 2, 3, 4, 6}},
	}

	for _, tt := range tests {
		actual := Union(tt.a, tt.b)
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("Union(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, actual)
		}
	}
}

func TestDifference(t *testing.T) {
	var tests = []struct {
		a        []int
		b        []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 4, 6}, []int{1, 3}},
	}

	for _, tt := range tests {
		actual := Difference(tt.a, tt.b)
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("Union(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, actual)
		}
	}
}
