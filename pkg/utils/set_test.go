package utils

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestIntersectInt(t *testing.T) {
	var tests = []struct {
		a        []int
		b        []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 4, 6}, []int{2}},
	}

	for _, tt := range tests {
		actual := IntersectInt(tt.a, tt.b)
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("IntersectInt(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, actual)
		}
	}
}

func TestUnionInt(t *testing.T) {
	var tests = []struct {
		a        []int
		b        []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 4, 6}, []int{1, 2, 3, 4, 6}},
	}

	for _, tt := range tests {
		actual := UnionInt(tt.a, tt.b)
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("UnionInt(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, actual)
		}
	}
}

func TestDifferenceInt(t *testing.T) {
	var tests = []struct {
		a        []int
		b        []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 4, 6}, []int{1, 3}},
	}

	for _, tt := range tests {
		actual := DifferenceInt(tt.a, tt.b)
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("DifferenceInt(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, actual)
		}
	}
}

func TestIntersectString(t *testing.T) {
	var tests = []struct {
		a        []string
		b        []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"b", "i", "n"}, []string{"b"}},
	}

	for _, tt := range tests {
		actual := IntersectString(tt.a, tt.b)
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("IntersectString(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, actual)
		}
	}
}

func TestUnionString(t *testing.T) {
	var tests = []struct {
		a        []string
		b        []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"b", "i", "n"}, []string{"a", "b", "c", "i", "n"}},
	}

	for _, tt := range tests {
		actual := UnionString(tt.a, tt.b)
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("UnionString(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, actual)
		}
	}
}

func TestDifferenceString(t *testing.T) {
	var tests = []struct {
		a        []string
		b        []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"b", "i", "n"}, []string{"a", "c"}},
	}

	for _, tt := range tests {
		actual := DifferenceString(tt.a, tt.b)
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("DifferenceString(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, actual)
		}
	}
}
