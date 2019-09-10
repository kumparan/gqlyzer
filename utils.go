package gqlyzer

import (
	"errors"
	"fmt"
)

func isNumber(c rune) bool {
	if c >= '0' && c <= '9' {
		return true
	}

	return false
}

func isAlphabet(c rune) bool {
	if (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c < 'Z') {
		return true
	}

	return false
}

func isWhitespace(c rune) bool {
	switch c {
	case '\n', '\t', ' ':
		return true
	default:
		return false
	}
}

func (l *Lexer) printParseStack() {
	for _, c := range l.parseStack {
		fmt.Print(string(c))
	}
	fmt.Println()
}

func (l *Lexer) isEOF() bool {
	return l.cursor >= len(l.input)
}

func (l *Lexer) read() (c rune, err error) {
	if l.isEOF() {
		err = errors.New("end of file")
	} else {
		c = rune(l.input[l.cursor])
	}
	return
}
