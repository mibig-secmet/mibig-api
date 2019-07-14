package queries

import (
	"encoding/json"
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

func TestQueryFromJson(t *testing.T) {
	var queryTests = []struct {
		input    []byte
		expected Query
	}{
		{[]byte(`{"search":"cluster","return_type":"json","terms":{"term_type":"expr","category":"type","term":"nrps"}}`), Query{Terms: &Expression{Category: "type", Term: "nrps"}}},
	}

	for _, tt := range queryTests {
		actual := Query{}
		err := json.Unmarshal(tt.input, &actual)
		if err != nil {
			t.Error(err)
		}
		if !cmp.Equal(tt.expected, actual) {
			t.Errorf("Unexpected Query from json (%s):\n%s", string(tt.input), cmp.Diff(tt.expected, actual))
		}
	}
}

func TestQueryToJson(t *testing.T) {
	var queryTests = []struct {
		expected []byte
		input    Query
	}{
		{[]byte(`{"search":"cluster","return_type":"json","terms":{"term_type":"expr","category":"type","term":"nrps"}}`), Query{Terms: &Expression{Category: "type", Term: "nrps"}}},
	}

	for _, tt := range queryTests {
		actual, err := json.Marshal(&tt.input)
		if err != nil {
			t.Error(err)
		}
		if !cmp.Equal(tt.expected, actual) {
			t.Errorf("Unexpected Query(%v) to json:\n%s", tt.input, cmp.Diff(tt.expected, actual))
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

func TestExpressionMarshalJson(t *testing.T) {
	var jsonTests = []struct {
		expr     Expression
		expected string
	}{
		{expr: Expression{Term: "nrps", Category: "type"}, expected: `{"term_type":"expr","category":"type","term":"nrps"}`},
	}

	for _, tt := range jsonTests {
		actual, err := json.Marshal(&tt.expr)
		if err != nil {
			t.Error(err)
		}
		if string(actual) != tt.expected {
			t.Errorf("Expression %v JSON marshalling unexpected: expected '%s', got '%s'", tt.expr, tt.expected, string(actual))
		}
	}
}

func TestExpressionUnmarshalJson(t *testing.T) {
	var jsonTests = []struct {
		expected Expression
		input    []byte
	}{
		{expected: Expression{Term: "nrps", Category: "type"}, input: []byte(`{"term_type":"expr","category":"type","term":"nrps"}`)},
	}

	for _, tt := range jsonTests {
		actual := Expression{}
		err := json.Unmarshal(tt.input, &actual)
		if err != nil {
			t.Error(err)
		}
		if !cmp.Equal(tt.expected, actual) {
			t.Errorf("Expression.Unmarshal(%s) unexpected result: %s", string(tt.input), cmp.Diff(tt.expected, actual))
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

func TestOperationMarshalJson(t *testing.T) {
	var jsonTests = []struct {
		op       Operation
		expected string
	}{
		{Operation{Operation: AND,
			Left:  &Expression{Category: "type", Term: "nrps"},
			Right: &Expression{Category: "genus", Term: "streptomyces"},
		}, `{"term_type":"op","operation":"and","left":{"term_type":"expr","category":"type","term":"nrps"},"right":{"term_type":"expr","category":"genus","term":"streptomyces"}}`},
		{Operation{Operation: OR,
			Left:  &Expression{Category: "type", Term: "nrps"},
			Right: &Expression{Category: "genus", Term: "streptomyces"},
		}, `{"term_type":"op","operation":"or","left":{"term_type":"expr","category":"type","term":"nrps"},"right":{"term_type":"expr","category":"genus","term":"streptomyces"}}`},
		{Operation{Operation: EXCEPT,
			Left:  &Expression{Category: "type", Term: "nrps"},
			Right: &Expression{Category: "genus", Term: "streptomyces"},
		}, `{"term_type":"op","operation":"except","left":{"term_type":"expr","category":"type","term":"nrps"},"right":{"term_type":"expr","category":"genus","term":"streptomyces"}}`},
		{Operation{Operation: AND,
			Left: &Expression{Term: "lanthipeptide", Category: "type"},
			Right: &Operation{Operation: OR,
				Left: &Operation{Operation: AND,
					Left:  &Expression{Term: "Streptomyces", Category: "genus"},
					Right: &Expression{Term: "coelicolor", Category: "species"},
				},
				Right: &Operation{Operation: AND,
					Left:  &Expression{Term: "Lactococcus", Category: "genus"},
					Right: &Expression{Term: "lactis", Category: "species"},
				},
			},
		}, `{"term_type":"op","operation":"and","left":{"term_type":"expr","category":"type","term":"lanthipeptide"},"right":{"term_type":"op","operation":"or","left":{"term_type":"op","operation":"and","left":{"term_type":"expr","category":"genus","term":"Streptomyces"},"right":{"term_type":"expr","category":"species","term":"coelicolor"}},"right":{"term_type":"op","operation":"and","left":{"term_type":"expr","category":"genus","term":"Lactococcus"},"right":{"term_type":"expr","category":"species","term":"lactis"}}}}`},
	}

	for _, tt := range jsonTests {
		actual, err := json.Marshal(&tt.op)
		if err != nil {
			t.Error(err)
		}
		if string(actual) != tt.expected {
			t.Errorf("Operation %v JSON marshalling unexpected: expected '%s', got '%s'", tt.op, tt.expected, string(actual))
		}
	}
}

func TestOperationUnmarshalJson(t *testing.T) {
	var jsonTests = []struct {
		expected Operation
		input    []byte
	}{
		{Operation{Operation: AND,
			Left:  &Expression{Category: "type", Term: "nrps"},
			Right: &Expression{Category: "genus", Term: "streptomyces"},
		}, []byte(`{"term_type":"op","operation":"and","left":{"term_type":"expr","category":"type","term":"nrps"},"right":{"term_type":"expr","category":"genus","term":"streptomyces"}}`)},
		{Operation{Operation: OR,
			Left:  &Expression{Category: "type", Term: "nrps"},
			Right: &Expression{Category: "genus", Term: "streptomyces"},
		}, []byte(`{"term_type":"op","operation":"or","left":{"term_type":"expr","category":"type","term":"nrps"},"right":{"term_type":"expr","category":"genus","term":"streptomyces"}}`)},
		{Operation{Operation: EXCEPT,
			Left:  &Expression{Category: "type", Term: "nrps"},
			Right: &Expression{Category: "genus", Term: "streptomyces"},
		}, []byte(`{"term_type":"op","operation":"except","left":{"term_type":"expr","category":"type","term":"nrps"},"right":{"term_type":"expr","category":"genus","term":"streptomyces"}}`)},
		{Operation{Operation: AND,
			Left: &Expression{Term: "lanthipeptide", Category: "type"},
			Right: &Operation{Operation: OR,
				Left: &Operation{Operation: AND,
					Left:  &Expression{Term: "Streptomyces", Category: "genus"},
					Right: &Expression{Term: "coelicolor", Category: "species"},
				},
				Right: &Operation{Operation: AND,
					Left:  &Expression{Term: "Lactococcus", Category: "genus"},
					Right: &Expression{Term: "lactis", Category: "species"},
				},
			},
		}, []byte(`{"term_type":"op","operation":"and","left":{"term_type":"expr","category":"type","term":"lanthipeptide"},"right":{"term_type":"op","operation":"or","left":{"term_type":"op","operation":"and","left":{"term_type":"expr","category":"genus","term":"Streptomyces"},"right":{"term_type":"expr","category":"species","term":"coelicolor"}},"right":{"term_type":"op","operation":"and","left":{"term_type":"expr","category":"genus","term":"Lactococcus"},"right":{"term_type":"expr","category":"species","term":"lactis"}}}}`)},
	}

	for _, tt := range jsonTests {
		actual := Operation{}
		err := json.Unmarshal(tt.input, &actual)
		if err != nil {
			t.Error(err)
		}
		if !cmp.Equal(tt.expected, actual) {
			t.Errorf("Expression.Unmarshal(%s) unexpected result: %s", string(tt.input), cmp.Diff(tt.expected, actual))
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
