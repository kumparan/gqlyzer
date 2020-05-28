package gqlyzer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kumparan/gqlyzer/token/operation"

	"github.com/kumparan/gqlyzer/token"
)

// Lexer definition
type Lexer struct {
	input      string
	parseStack []rune
	cursor     int
}

// New use to init lexer
func New(gql string) (l *Lexer) {
	l = &Lexer{
		input: gql,
	}
	l.Reset()

	return l
}

// Reset reset the state of lexer
func (l *Lexer) Reset() {
	l.parseStack = []rune{}
	l.cursor = 0
}

// Parse operation without variable
func (l *Lexer) Parse() (token.Operation, error) {
	return l.parseOperation()
}

// ParseOperationType parse operation type only
func (l *Lexer) ParseOperationType() (operation.Type, error) {
	ot, _, err := l.parseOperationType()
	return ot, err
}

// ParseWithVariables parse operation with variable
func (l *Lexer) ParseWithVariables(variables string) (token.Operation, error) {
	variableMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(variables), &variableMap)
	if err != nil {
		return token.Operation{}, err
	}

	for key, content := range variableMap {
		var s string
		switch c := content.(type) {
		case string:
			s = fmt.Sprintf("\"%s\"", c)
		case int:
			s = string(c)
		default:
			jsonStr, err := json.Marshal(content)
			if err != nil {
				return token.Operation{}, err
			}
			s = string(jsonStr)
		}
		l.input = strings.ReplaceAll(l.input, "$"+key, s)
	}

	return l.parseOperation()
}
