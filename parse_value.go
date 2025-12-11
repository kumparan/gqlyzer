package gqlyzer

import "errors"

// temporary immplementation to parse value such as number, enum etc
func (l *Lexer) parseOtherAsString() (value string, err error) {
	l.pushFlush()
	c, err := l.read()
	for err == nil &&
		!isWhitespace(c) &&
		c != ')' &&
		c != ',' &&
		c != '}' {
		l.push(c)
		l.cursor++
		c, err = l.read()
	}

	return l.popFlush()
}

func (l *Lexer) parseString() (value string, err error) {
	c, err := l.read()
	if err != nil {
		return
	}

	if c == '\'' || c == '"' {
		l.push(c)
	} else {
		err = errors.New("value is  not a string")
		return
	}

	l.pushFlush()
	l.cursor++
	c, err = l.read()
	for err == nil &&
		!isWhitespace(c) &&
		c != '\'' &&
		c != '"' {
		l.push(c)
		l.cursor++
		c, err = l.read()
	}

	content, err := l.popFlush()
	if err != nil {
		return
	}

	if c != l.pop() {
		err = errors.New("no " + string(c) + " found")
		return
	}

	l.cursor++
	return `"` + content + `"`, nil
}

// TODO: handle if user's text input contains "]<whitespace>)"
func (l *Lexer) parseText() (value string, err error) {
	c, err := l.read()
	if err != nil {
		return
	}

	if c == '[' {
		l.push(c)
	} else {
		err = errors.New("value is not a text")
		return
	}

	l.pushFlush()

parseTextPart:
	l.cursor++
	c, err = l.read()
	for err == nil && c != ']' {
		l.push(c)
		l.cursor++
		c, err = l.read()
	}

	l.cursor++
	c, err = l.read()
	if err != nil {
		return
	}
	if c == '\n' { // currently any text payload that consists ]\n<whitespace><next char> will be parsed as ]<next char>
		l.consumeWhitespace()
	}

	c, err = l.read()
	if err != nil {
		return
	}
	if c != ')' {
		l.push(c)
		goto parseTextPart
	}

	content, err := l.popFlush()
	if err != nil {
		return
	}

	if l.pop() != '[' {
		err = errors.New("no " + string(c) + " found")
		return
	}

	return `[` + content + `]`, nil
}
