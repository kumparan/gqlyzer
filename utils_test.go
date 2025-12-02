package gqlyzer

import (
	"testing"
)

func TestIsAlphabet(t *testing.T) {

	t.Run("lowercase letters", func(t *testing.T) {
		for c := 'a'; c <= 'z'; c++ {
			if !isAlphabet(c) {
				t.Errorf("isAlphabet(%q) return false, expect true", c)
			}
		}
	})

	t.Run("uppercase letters", func(t *testing.T) {
		for c := 'A'; c <= 'Z'; c++ {
			if !isAlphabet(c) {
				t.Errorf("isAlphabet(%q) return false, expect true", c)
			}
		}
	})

	t.Run("unicode letters", func(t *testing.T) {
		unicodeLetters := []rune{'ø', '中', 'д'}

		for _, c := range unicodeLetters {
			if !isAlphabet(c) {
				t.Errorf("isAlphabet(%q) return false, expect true", c)
			}
		}
	})

	t.Run("non letters", func(t *testing.T) {
		nonLetters := []rune{'😔', '\n', '\t', ' '}

		for _, c := range nonLetters {
			if isAlphabet(c) {
				t.Errorf("isAlphabet(%q) return true, expect false", c)
			}
		}
	})
}
