package gqlyzer

import "github.com/kumparan/gqlyzer/token"

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
