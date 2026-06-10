package gqlyzer

import (
	"errors"
	"unicode"
)

// ErrEOF is returned by read() when the input is exhausted.
// It is exported so callers can distinguish a clean end-of-input
// from a real mid-parse failure using errors.Is.
var ErrEOF = errors.New("end of file")

func isNumber(c rune) bool {
	if c >= '0' && c <= '9' {
		return true
	}
	return false
}

func isAlphabet(c rune) bool {
	return unicode.IsLetter(c)
}

func isWhitespace(c rune) bool {
	return unicode.IsSpace(c)
}

func (l *Lexer) isEOF() bool {
	return l.cursor >= len(l.input)
}

func (l *Lexer) read() (c rune, err error) {
	if l.isEOF() {
		err = ErrEOF
	} else {
		c = rune(l.input[l.cursor])
	}
	return
}

func (l *Lexer) consumeWhitespace() {
	c, err := l.read()
	for err == nil && isWhitespace(c) {
		if c == '\n' {
			l.push('\\')
		}
		l.cursor++
		c, err = l.read()
	}
}

// commented out since no usage, needed in development
//func (l *Lexer) printParseStack() {
//	for _, c := range l.parseStack {
//		fmt.Print(string(c))
//	}
//	fmt.Println()
//}
