package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
	// keep calling Parse on whatever wasn't parsed until you get done = true
	_, ok, _ := headers.Parse(data[n:])
	assert.True(t, ok)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	t.Run("valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "localhost", headers["host"])
		assert.Equal(t, len("Host: localhost\r\n"), n)
		assert.False(t, done)
	})

	t.Run("Non-valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("   Host   :   localhost   \r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["existing"] = "Thing"

		n1, done1, err1 := headers.Parse([]byte("Host: localhost\r\n"))
		require.NoError(t, err1)
		assert.False(t, done1)

		n2, done2, err2 := headers.Parse([]byte("User-Agent: test\r\n"))
		require.NoError(t, err2)
		assert.False(t, done2)

		assert.Equal(t, "Thing", headers["existing"])
		assert.Equal(t, "localhost", headers["host"])
		assert.Equal(t, "test", headers["user-agent"])
		assert.Equal(t, len("Host: localhost\r\n"), n1)
		assert.Equal(t, len("User-Agent: test\r\n"), n2)
	})

	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		n, done, err := headers.Parse([]byte("\r\n"))
		require.NoError(t, err)
		assert.True(t, done)
		assert.Equal(t, 2, n)
	})

	t.Run("Invalid spacing header (Host : foo)", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host : bad\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Valid header with uppercase field-name", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("ConTent-TyPe: text/html\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)

		assert.Equal(t, "text/html", headers["content-type"]) // lowercase expected
		assert.Equal(t, len("ConTent-TyPe: text/html\r\n"), n)
		assert.False(t, done)
	})

	t.Run("Invalid — illegal unicode character in field-name", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("H©st: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})
}
