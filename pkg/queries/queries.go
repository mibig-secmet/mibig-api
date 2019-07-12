package queries

import (
	"errors"
	"fmt"
	"strings"
)

type QueryTerm interface {
	Query() string
	/*
		MarshalJSON() ([]byte, error)
		UnmarshalJSON(data []byte) error
	*/
}

type QueryType int

const (
	Cluster QueryType = iota
	Cds
	Domain
)

type ReturnType int

const (
	Json ReturnType = iota
	Csv
	NucleotideFasta
	AminoAcidFasta
)

type Query struct {
	QueryType  QueryType  `json:"search"`
	ReturnType ReturnType `json:"return_type"`
	Terms      QueryTerm  `json:"terms"`
}

// TODO: Add validation
func NewQueryFromString(input string) (*Query, error) {
	var err error
	query := Query{QueryType: Cluster, ReturnType: Json}
	parser := NewParser(input)
	if query.Terms, err = getTerm(parser); err != nil {
		return nil, err
	}

	return &query, nil
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

type OperationType int

const (
	AND OperationType = iota
	OR
	EXCEPT
)

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
	parser := Parser{
		keywords: map[string]OperationType{
			"and":    AND,
			"or":     OR,
			"except": EXCEPT,
		},
	}

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
