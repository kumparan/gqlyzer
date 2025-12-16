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

// TODO: handle query like user input, or contains ]
func (l *Lexer) parseArray() (value string, err error) {
	c, err := l.read()
	if err != nil {
		return
	}

	if c == '[' {
		l.push(c)
	} else {
		err = errors.New("value is not an array")
		return
	}

	l.pushFlush()
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

	content, err := l.popFlush()
	if err != nil {
		return
	}

	if l.pop() != '[' {
		err = errors.New("no " + string(c) + " found")
		return
	}

	l.cursor++
	return `[` + content + `]`, nil
}
