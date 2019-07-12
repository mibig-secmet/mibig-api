package queries

import (
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

func TestQueryFromString(t *testing.T) {
	var tests = []struct {
		input    string
		err      string
		expected *Query
	}{
		{"", "Unexpected end of expression", nil},
		{"nrps", "", &Query{QueryType: Cluster, ReturnType: Json,
			Terms: &Expression{Term: "nrps", Category: "unknown"}},
		},
		{"[type]nrps", "", &Query{QueryType: Cluster, ReturnType: Json,
			Terms: &Expression{Term: "nrps", Category: "type"}},
		},
		{"nrps AND 1234", "", &Query{QueryType: Cluster, ReturnType: Json,
			Terms: &Operation{Operation: AND,
				Left:  &Expression{Term: "nrps", Category: "unknown"},
				Right: &Expression{Term: "1234", Category: "unknown"}},
		}},
		{"nrps OR 1234", "", &Query{QueryType: Cluster, ReturnType: Json,
			Terms: &Operation{Operation: OR,
				Left:  &Expression{Term: "nrps", Category: "unknown"},
				Right: &Expression{Term: "1234", Category: "unknown"}},
		}},
		{"nrps EXCEPT 1234", "", &Query{QueryType: Cluster, ReturnType: Json,
			Terms: &Operation{Operation: EXCEPT,
				Left:  &Expression{Term: "nrps", Category: "unknown"},
				Right: &Expression{Term: "1234", Category: "unknown"}},
		}},
		{"nrps 1234", "", &Query{QueryType: Cluster, ReturnType: Json,
			Terms: &Operation{Operation: AND,
				Left:  &Expression{Term: "nrps", Category: "unknown"},
				Right: &Expression{Term: "1234", Category: "unknown"}},
		}},
		{"ripp AND ( streptomyces OR lactococcus )", "", &Query{QueryType: Cluster, ReturnType: Json,
			Terms: &Operation{Operation: AND,
				Left: &Expression{Term: "ripp", Category: "unknown"},
				Right: &Operation{Operation: OR,
					Left:  &Expression{Term: "streptomyces", Category: "unknown"},
					Right: &Expression{Term: "lactococcus", Category: "unknown"},
				}},
		}},
		{"lanthipeptide ((Streptomyces coelicolor) OR (Lactococcus lactis))", "", &Query{QueryType: Cluster, ReturnType: Json,
			Terms: &Operation{Operation: AND,
				Left: &Expression{Term: "lanthipeptide", Category: "unknown"},
				Right: &Operation{Operation: OR,
					Left: &Operation{Operation: AND,
						Left:  &Expression{Term: "Streptomyces", Category: "unknown"},
						Right: &Expression{Term: "coelicolor", Category: "unknown"},
					},
					Right: &Operation{Operation: AND,
						Left:  &Expression{Term: "Lactococcus", Category: "unknown"},
						Right: &Expression{Term: "lactis", Category: "unknown"},
					},
				},
			},
		}},
		{"AND ripp", "Invalid use of keyword AND", nil},
		{"( ripp", "Invalid token END", nil},
		{"END", "Malformatted input", nil},
	}

	for _, tt := range tests {
		actual, err := NewQueryFromString(tt.input)
		if !ErrorContains(err, tt.err) {
			t.Errorf("NewQueryFromString(%s) unexpected error. Expected %s, got %s", tt.input, tt.err, err)
		}
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("NewQueryFromString(%s) differs from expected:\n%s", tt.input, cmp.Diff(actual, tt.expected))
		}
	}
}

func TestGenerateTokens(t *testing.T) {
	var tokenTests = []struct {
		input    string
		expected []string
	}{
		{"foo", []string{"foo", "END"}},
		{"(foo)", []string{"(", "foo", ")", "END"}},
		{"foo (foo) foo", []string{"foo", "(", "foo", ")", "foo", "END"}},
	}

	for _, tt := range tokenTests {
		actual := generateTokens(tt.input)
		if !cmp.Equal(actual, tt.expected) {
			t.Errorf("generateTokens(%s): expected %v, got %v", tt.input, tt.expected, actual)
		}
	}
}

func TestExpressionQuery(t *testing.T) {
	var queryTests = []struct {
		expr     Expression
		expected string
	}{
		{Expression{Category: "type", Term: "nrps"}, "[type]nrps"},
		{Expression{Category: "unknown", Term: "nrps"}, "nrps"},
	}

	for _, tt := range queryTests {
		actual := tt.expr.Query()
		if actual != tt.expected {
			t.Errorf("Expression.Query(%v): expected '%s', got '%s'", tt.expr, tt.expected, actual)
		}
	}
}

func TestOperationQuery(t *testing.T) {
	var queryTests = []struct {
		op       Operation
		expected string
	}{
		{Operation{Operation: AND,
			Left:  &Expression{Category: "type", Term: "nrps"},
			Right: &Expression{Category: "unknown", Term: "nrps"},
		}, "( [type]nrps AND nrps )"},
		{Operation{Operation: OR,
			Left:  &Expression{Category: "type", Term: "nrps"},
			Right: &Expression{Category: "unknown", Term: "nrps"},
		}, "( [type]nrps OR nrps )"},
		{Operation{Operation: EXCEPT,
			Left:  &Expression{Category: "type", Term: "nrps"},
			Right: &Expression{Category: "unknown", Term: "nrps"},
		}, "( [type]nrps EXCEPT nrps )"},
	}

	for _, tt := range queryTests {
		actual := tt.op.Query()
		if actual != tt.expected {
			t.Errorf("Expression.Query(%v): expected '%s', got '%s'", tt.op, tt.expected, actual)
		}
	}
}

func ErrorContains(out error, want string) bool {
	if out == nil {
		return want == ""
	}
	if want == "" {
		return false
	}
	return strings.Contains(out.Error(), want)
}
