package gqlyzer

import (
	"errors"

	"github.com/kumparan/gqlyzer/token"
)

func (l *Lexer) parseArgument() (argument token.Argument, err error) {
	if x := l.pop(); x != ',' && x != '\\' {
		err = errors.New("expected separator")
		return
	}

	name, err := l.parseName()
	if err != nil {
		return
	}
	argument.Key = name

	l.consumeWhitespace()
	c, err := l.read()
	if err != nil {
		return
	}
	if c != ':' {
		err = errors.New("expected : but found " + string(c))
		return
	}

	l.cursor++
	l.consumeWhitespace()
	_, err = l.read()
	if err != nil {
		return
	}
	if subArg, err := l.parseArgumentValueObject(); err == nil {
		argument.ObjectValue = subArg
	} else if value, err := l.parseString(); err == nil {
		argument.Value = value
	} else {
		value, err := l.parseOtherAsString()
		if err != nil {
			err = errors.New("cannot parse value")
			return token.Argument{}, err
		}
		argument.Value = value
	}

	return
}

func (l *Lexer) parseArgumentSet() (set token.ArgumentSet, err error) {
	set = make(token.ArgumentSet)
	l.consumeWhitespace()
	c, err := l.read()
	if err != nil {
		return
	}
	if c == '(' {
		l.push('(')
		l.pushFlush()
		l.push('\\')
		if err != nil {
			return
		}
		l.cursor++
		l.consumeWhitespace()
		c, err = l.read()
		for err == nil && c != ')' {
			if c == ',' {
				l.push(c)
				l.cursor++
				l.consumeWhitespace()
				c, err = l.read()
				continue
			}

			arg, err := l.parseArgument()
			if err != nil {
				return token.ArgumentSet{}, err
			}
			set[arg.Key] = arg
			l.consumeWhitespace()
			c, err = l.read()
			if err != nil {
				return token.ArgumentSet{}, err
			}
		}
		_, err = l.popFlush()
		if err != nil {
			return
		}

		err = l.popCond('(')
		if err != nil {
			return
		}
	} else {
		return token.ArgumentSet{}, nil
	}

	return
}

// parse value of  an  argument if  the value is an object
func (l *Lexer) parseArgumentValueObject() (set token.ArgumentSet, err error) {
	set = make(token.ArgumentSet)
	l.consumeWhitespace()
	c, err := l.read()
	if err != nil {
		return
	}

	if c == '{' {
		l.push('{')
		l.pushFlush()
		if err != nil {
			return
		}
		l.cursor++
		l.consumeWhitespace()
		c, err = l.read()
		for err == nil && c != '}' {
			if c == ',' {
				l.push(c)
				l.cursor++
				l.consumeWhitespace()
				c, err = l.read()
				continue
			}

			arg, err := l.parseArgument()
			if err != nil {
				return token.ArgumentSet{}, err
			}
			set[arg.Key] = arg
			l.cursor++
			l.consumeWhitespace()
			c, err = l.read()
			if err != nil {
				return token.ArgumentSet{}, err
			}
		}
		_, err = l.popFlush()
		if err != nil {
			return
		}

		err = l.popCond('{')
		if err != nil {
			return
		}
		l.cursor++
		l.consumeWhitespace()
	} else {
		return token.ArgumentSet{}, errors.New("argument is not an object")
	}

	return
}
