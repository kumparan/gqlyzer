package gqlyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgument(t *testing.T) {
	t.Run("without string value", func(t *testing.T) {
		l := Lexer{input: `SomeQuery: "helloworld"`}
		l.Reset()
		l.push('\\')

		s, err := l.parseArgument()
		assert.NoError(t, err)
		assert.Equal(t, "SomeQuery", s.Key)
		assert.Equal(t, `"helloworld"`, s.Value)
	})

	t.Run("without object value", func(t *testing.T) {
		l := Lexer{input: `SomeQuery: {
			test: "helloworld",
			test2: "helloworld"	
		}`}
		l.Reset()
		l.push('\\')

		s, err := l.parseArgument()
		assert.NoError(t, err)
		assert.Equal(t, "SomeQuery", s.Key)
		assert.Equal(t, "test", s.ObjectValue["test"].Key)
		assert.Equal(t, `"helloworld"`, s.ObjectValue["test"].Value)
		assert.Equal(t, "test2", s.ObjectValue["test2"].Key)
	})

}

func TestParseArgumentSet(t *testing.T) {
	t.Run("with single value", func(t *testing.T) {
		l := Lexer{input: `( 
		arg1: 1, 
		arg2: 2,
		arg3: 3
 )`}
		l.Reset()

		s, err := l.parseArgumentSet()

		assert.NoError(t, err)
		assert.Equal(t, "arg1", s["arg1"].Key)
		assert.Equal(t, "1", s["arg1"].Value)
		assert.Equal(t, "arg2", s["arg2"].Key)
		assert.Equal(t, "2", s["arg2"].Value)
		assert.Equal(t, "arg3", s["arg3"].Key)
		assert.Equal(t, "3", s["arg3"].Value)
	})

	t.Run("with nested value", func(t *testing.T) {
		l := Lexer{input: `(
			user: {
				name: "danu",
				id: 123
			}, 
			page: 1
		)`}
		l.Reset()

		s, err := l.parseArgumentSet()

		assert.NoError(t, err)
		assert.Equal(t, "user", s["user"].Key)
		assert.Equal(t, "page", s["page"].Key)
		assert.Equal(t, "1", s["page"].Value)
		assert.Equal(t, `"danu"`, s["user"].ObjectValue["name"].Value)
		assert.Equal(t, "123", s["user"].ObjectValue["id"].Value)
	})

}
