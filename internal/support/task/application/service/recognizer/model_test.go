package recognizer
import (
	"testing"
	"github.com/stretchr/testify/assert"
)
func TestRecognized(t *testing.T) {
	t.Run("Store recognized text", func(t *testing.T) {
		result := Recognized{Text: "text"}
		assert.Equal(t, "text", result.Text, "unexpected recognized text")
	})
}