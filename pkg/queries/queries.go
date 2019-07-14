package queries

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type QueryTerm interface {
	Query() string
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

type QueryType int

const (
	Cluster QueryType = iota
	Cds
	Domain
)

var STRING_QUERY_TYPE_MAP = map[string]QueryType{
	"cluster": Cluster,
	"cds":     Cds,
	"domain":  Domain,
}

var QUERY_TYPE_STRING_MAP = map[QueryType]string{
	Cluster: "cluster",
	Cds:     "cds",
	Domain:  "domain",
}

type ReturnType int

const (
	Json ReturnType = iota
	Csv
	NucleotideFasta
	AminoAcidFasta
)

var STRING_RETURN_TYPE_MAP = map[string]ReturnType{
	"json":   Json,
	"csv":    Csv,
	"fasta":  NucleotideFasta,
	"fastaa": AminoAcidFasta,
}

var RETURN_TYPE_STRING_MAP = map[ReturnType]string{
	Json:            "json",
	Csv:             "csv",
	NucleotideFasta: "fasta",
	AminoAcidFasta:  "fastaa",
}

type Query struct {
	QueryType  QueryType  `json:"search"`
	ReturnType ReturnType `json:"return_type"`
	Terms      QueryTerm  `json:"terms"`
}

func NewQueryFromString(input string) (*Query, error) {
	var err error
	query := Query{QueryType: Cluster, ReturnType: Json}
	parser := NewParser(input)
	if query.Terms, err = getTerm(parser); err != nil {
		return nil, err
	}

	return &query, nil
}

func (q *Query) MarshalJSON() ([]byte, error) {
	var tmp struct {
		QueryString  string          `json:"search"`
		ReturnString string          `json:"return_type"`
		Terms        json.RawMessage `json:"terms"`
	}
	var err error
	var ok bool

	tmp.Terms, err = json.Marshal(q.Terms)
	if err != nil {
		return nil, err
	}

	tmp.QueryString, ok = QUERY_TYPE_STRING_MAP[q.QueryType]
	if !ok {
		return nil, fmt.Errorf("Unexpected QueryType %d", q.QueryType)
	}

	tmp.ReturnString, ok = RETURN_TYPE_STRING_MAP[q.ReturnType]
	if !ok {
		return nil, fmt.Errorf("Unexpected ReturnType %d", q.ReturnType)
	}

	return json.Marshal(&tmp)
}

func (q *Query) UnmarshalJSON(data []byte) error {
	var tmp struct {
		QueryString  string          `json:"search"`
		ReturnString string          `json:"return_type"`
		Terms        json.RawMessage `json:"terms"`
	}
	var ok bool

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	q.QueryType, ok = STRING_QUERY_TYPE_MAP[strings.ToLower(tmp.QueryString)]
	if !ok {
		return fmt.Errorf("Invalid query type %s", tmp.QueryString)
	}

	q.ReturnType, ok = STRING_RETURN_TYPE_MAP[strings.ToLower(tmp.ReturnString)]
	if !ok {
		return fmt.Errorf("Invalid return type %s", tmp.ReturnString)
	}

	q.Terms, err = unmarshalTerm(tmp.Terms)
	if err != nil {
		return err
	}

	return nil
}

type Expression struct {
	Category string `json:"category"`
	Term     string `json:"term"`
}

func (e *Expression) Query() string {
	category := ""
	if e.Category != "unknown" {
		category = fmt.Sprintf("[%s]", e.Category)
	}
	return fmt.Sprintf("%s%s", category, e.Term)
}

func (e *Expression) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     string `json:"term_type"`
		Category string `json:"category"`
		Term     string `json:"term"`
	}{Type: "expr", Category: e.Category, Term: e.Term})
}

func (e *Expression) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Category string `json:"category"`
		Term     string `json:"term"`
	}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	e.Category = tmp.Category
	e.Term = tmp.Term

	return nil
}

type Operation struct {
	Operation OperationType `json:"operation"`
	Left      QueryTerm     `json:"left"`
	Right     QueryTerm     `json:"right"`
}

func (o *Operation) Op() string {
	switch op := o.Operation; op {
	case AND:
		return "AND"
	case OR:
		return "OR"
	case EXCEPT:
		return "EXCEPT"
	}
	return "INVALID"
}

func (o *Operation) Query() string {
	return fmt.Sprintf("( %s %s %s )", o.Left.Query(), o.Op(), o.Right.Query())
}

func (o *Operation) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      string    `json:"term_type"`
		Operation string    `json:"operation"`
		Left      QueryTerm `json:"left"`
		Right     QueryTerm `json:"right"`
	}{Type: "op", Operation: strings.ToLower(o.Op()), Left: o.Left, Right: o.Right})
}

func (o *Operation) UnmarshalJSON(data []byte) error {
	var tmp OperationEnvelope
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	var ok bool
	o.Operation, ok = STRING_OP_MAP[strings.ToLower(tmp.Operation)]
	if !ok {
		return fmt.Errorf("Invalid operation '%s'", tmp.Operation)
	}
	o.Left, err = unmarshalTerm(tmp.Left)
	if err != nil {
		return err
	}
	o.Right, err = unmarshalTerm(tmp.Right)
	if err != nil {
		return err
	}
	return nil
}

func unmarshalTerm(data json.RawMessage) (QueryTerm, error) {
	var spy TermSpy

	err := json.Unmarshal(data, &spy)
	if err != nil {
		return nil, err
	}

	switch term_type := strings.ToLower(spy.Type); term_type {
	case "expr":
		{
			var expr Expression
			err := json.Unmarshal(data, &expr)
			if err != nil {
				return nil, err
			}
			return &expr, nil
		}
	case "op":
		{
			var op Operation
			err := json.Unmarshal(data, &op)
			if err != nil {
				return nil, err
			}
			return &op, nil
		}
	}

	return nil, fmt.Errorf("Invalid term_type '%s'", spy.Type)
}

type TermSpy struct {
	Type string `json:"term_type"`
}

type OperationEnvelope struct {
	Operation string          `json:"operation"`
	Left      json.RawMessage `json:"left"`
	Right     json.RawMessage `json:"right"`
}

type OperationType int

const (
	AND OperationType = iota
	OR
	EXCEPT
)

var STRING_OP_MAP = map[string]OperationType{
	"and":    AND,
	"or":     OR,
	"except": EXCEPT,
}

type Parser struct {
	tokens   []string
	keywords map[string]OperationType
}

func (p *Parser) Peek() string {
	return p.tokens[0]
}

func (p *Parser) Consume() string {
	token := p.tokens[0]
	p.tokens = p.tokens[1:]
	return token
}

func (p *Parser) ConsumeExpected(expected string) bool {
	token := p.Peek()
	if token == expected {
		p.Consume()
		return true
	}
	return false
}

func NewParser(input string) *Parser {
	parser := Parser{keywords: STRING_OP_MAP}

	parser.tokens = generateTokens(input)

	return &parser
}

func getTerm(parser *Parser) (QueryTerm, error) {
	if len(parser.tokens) < 2 {
		return nil, errors.New("Unexpected end of expression")
	}

	left, err := getExpression(parser)
	if err != nil {
		return nil, err
	}

	next_token := parser.Peek()
	if op, ok := parser.keywords[strings.ToLower(next_token)]; ok {
		parser.Consume()
		right, err := getExpression(parser)
		if err != nil {
			return nil, err
		}
		return &Operation{Operation: op, Left: left, Right: right}, nil
	}
	if next_token == "END" || next_token == ")" {
		return left, nil
	}
	// Two expressions without keyword will be ANDed

	right, err := getExpression(parser)
	if err != nil {
		return nil, err
	}
	return &Operation{Operation: AND, Left: left, Right: right}, nil
}

func getExpression(parser *Parser) (QueryTerm, error) {
	if parser.ConsumeExpected("(") {
		term, err := getTerm(parser)
		if err != nil {
			return nil, err
		}
		if !parser.ConsumeExpected(")") {
			return nil, fmt.Errorf("Invalid token %s", parser.Peek())
		}
		return term, nil
	}

	raw_expression := parser.Consume()
	if _, ok := parser.keywords[strings.ToLower(raw_expression)]; ok {
		return nil, fmt.Errorf("Invalid use of keyword %s", raw_expression)
	}
	if raw_expression == "END" {
		return nil, fmt.Errorf("Malformatted input")
	}
	category := "unknown"
	term := raw_expression

	if strings.HasPrefix(raw_expression, "[") {
		end := strings.Index(raw_expression, "]")
		if end > -1 {
			category = raw_expression[1:end]
			term = raw_expression[end+1:]
		}
	}

	return &Expression{Term: term, Category: category}, nil
}

func generateTokens(input string) []string {
	final_tokens := []string{}

	tokens := strings.Fields(strings.ReplaceAll(strings.ReplaceAll(input, ")", " ) "), "(", " ( "))

	for _, t := range tokens {
		final_tokens = append(final_tokens, t)
	}

	final_tokens = append(final_tokens, "END")

	return final_tokens
}
