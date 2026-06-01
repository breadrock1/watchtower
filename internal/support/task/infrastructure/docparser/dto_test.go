package docparser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testParsedText       = "parsed"
	testEmptyText        = ""
	testTextWithSpaces   = "  parsed text  "
	testMultilineText    = "first line\nsecond line\nthird line"
	testUnicodeText      = "Привет, документ распознан"
	testSpecialCharsText = "price:\t1000\nstatus: \"ok\""
)

func TestParsedContent(t *testing.T) {
	t.Run("Convert to recognized", func(t *testing.T) {
		got := (&ParsedContent{Text: testParsedText}).ToRecognized()

		assert.Equal(t, testParsedText, got.Text, "unexpected recognized content")
	})

	t.Run("Convert empty text", func(t *testing.T) {
		got := (&ParsedContent{Text: testEmptyText}).ToRecognized()

		assert.Equal(t, testEmptyText, got.Text, "expected empty recognized text")
	})

	t.Run("Keep text spaces", func(t *testing.T) {
		got := (&ParsedContent{Text: testTextWithSpaces}).ToRecognized()

		assert.Equal(t, testTextWithSpaces, got.Text, "expected spaces to be preserved")
	})

	t.Run("Keep multiline text", func(t *testing.T) {
		got := (&ParsedContent{Text: testMultilineText}).ToRecognized()

		assert.Equal(t, testMultilineText, got.Text, "expected multiline text to be preserved")
	})

	t.Run("Keep unicode text", func(t *testing.T) {
		got := (&ParsedContent{Text: testUnicodeText}).ToRecognized()

		assert.Equal(t, testUnicodeText, got.Text, "expected unicode text to be preserved")
	})

	t.Run("Keep special characters", func(t *testing.T) {
		got := (&ParsedContent{Text: testSpecialCharsText}).ToRecognized()

		assert.Equal(t, testSpecialCharsText, got.Text, "expected special characters to be preserved")
	})

	t.Run("Panic on nil parsed content", func(t *testing.T) {
		var content *ParsedContent

		assert.Panics(t, func() {
			content.ToRecognized()
		}, "expected panic for nil parsed content")
	})
}
