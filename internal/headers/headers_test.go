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
	assert.Equal(t, "localhost:42069", (*headers)["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("Content-Type:   application/json   \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "application/json", (*headers)["content-type"])
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	(*headers)["existing"] = "header"
	data = []byte("Host: localhost:42069\r\nContent-Type: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", (*headers)["host"])
	assert.Equal(t, "application/json", (*headers)["content-type"])
	assert.Equal(t, "header", (*headers)["existing"])
	assert.Equal(t, 55, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character in header key
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Existing header gets overwritten by parsed data
	headers = NewHeaders()
	(*headers)["host"] = "oldhost:8080"
	data = []byte("Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "oldhost:8080, localhost:42069", (*headers)["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestStandardHeaders(t *testing.T) {
	// Test: Two standard HTTP headers parsed correctly
	headers := NewHeaders()
	data := []byte("Host: example.com\r\nContent-Type: text/html\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "example.com", (*headers)["host"])
	assert.Equal(t, "text/html", (*headers)["content-type"])
	assert.Equal(t, 44, n)
	assert.False(t, done)

	// Test: Headers with special characters in values
	headers = NewHeaders()
	data = []byte("Authorization: Bearer abc123!@#$%^&*()\r\nX-Custom-Header: value=with;special:chars\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "Bearer abc123!@#$%^&*()", (*headers)["authorization"])
	assert.Equal(t, "value=with;special:chars", (*headers)["x-custom-header"])
	assert.False(t, done)

	// Test: Headers with numeric values
	headers = NewHeaders()
	data = []byte("X-Request-ID: 12345\r\nX-Rate-Limit: 100\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "12345", (*headers)["x-request-id"])
	assert.Equal(t, "100", (*headers)["x-rate-limit"])
	assert.False(t, done)

	// Test: Header with hyphenated name
	headers = NewHeaders()
	data = []byte("X-Custom-Long-Header-Name: some-value\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "some-value", (*headers)["x-custom-long-header-name"])
	assert.False(t, done)
}

func TestEmptyHeaders(t *testing.T) {
	// Test: Empty header value
	headers := NewHeaders()
	data := []byte("X-Empty-Header:\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "", (*headers)["x-empty-header"])
	assert.Equal(t, 17, n)
	assert.False(t, done)

	// Test: Header value with only whitespace (should be trimmed to empty)
	headers = NewHeaders()
	data = []byte("X-Whitespace:     \r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "", (*headers)["x-whitespace"])
	assert.False(t, done)

	// Test: Just the end of headers marker (no headers)
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)
	assert.Equal(t, 0, len(*headers))
}

func TestMalformedHeader(t *testing.T) {
	// Test: Header without colon separator
	headers := NewHeaders()
	data := []byte("InvalidHeaderNoColon\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, ErrHeaderFieldsSeperator, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header with space before colon (space within header name)
	headers = NewHeaders()
	data = []byte("Invalid Header: value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	// Space in header name causes it to be invalid
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header key with invalid unicode character
	headers = NewHeaders()
	data = []byte("Inválid: value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, ErrHeaderKeyInvalidChar, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header key with control character
	headers = NewHeaders()
	data = []byte("Header\x00Name: value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, ErrHeaderKeyInvalidChar, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header key starting with whitespace
	// Note: Leading whitespace is trimmed, so " Content-Type" becomes "content-type"
	// However, it first checks for " :" pattern which would fail
	headers = NewHeaders()
	data = []byte(" Content-Type: text/html\r\n\r\n")
	_, done, err = headers.Parse(data)
	// The implementation trims whitespace, so this actually parses successfully
	// as "content-type" with value "text/html"
	require.NoError(t, err)
	assert.Equal(t, "text/html", (*headers)["content-type"])
	assert.False(t, done)

	// Test: Multiple colons in header (should be valid - only split on first colon)
	headers = NewHeaders()
	data = []byte("X-Time: 10:30:45\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "10:30:45", (*headers)["x-time"])
	assert.False(t, done)

	// Test: Header with tab character in value (should be valid)
	headers = NewHeaders()
	data = []byte("X-Tab-Value: value\twith\ttabs\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "value\twith\ttabs", (*headers)["x-tab-value"])
	assert.False(t, done)

	// Test: Header with empty key (just colon)
	headers = NewHeaders()
	data = []byte(": value\r\n\r\n")
	_, done, err = headers.Parse(data)
	// Empty key after trimming - behavior depends on implementation
	require.NoError(t, err) // Empty string passes isValidHeaderKey
	assert.Equal(t, "value", (*headers)[""])
	assert.False(t, done)
}

func TestDuplicateHeaders(t *testing.T) {
	// Test: Same header appearing twice in the same parse call
	headers := NewHeaders()
	data := []byte("Set-Cookie: session=abc123\r\nSet-Cookie: tracking=xyz789\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "session=abc123, tracking=xyz789", (*headers)["set-cookie"])
	assert.Equal(t, 57, n)
	assert.False(t, done)

	// Test: Duplicate headers with different casing (should combine)
	headers = NewHeaders()
	data = []byte("Accept: text/html\r\nACCEPT: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "text/html, application/json", (*headers)["accept"])
	assert.Equal(t, 45, n)
	assert.False(t, done)

	// Test: Pre-existing header combined with parsed header
	headers = NewHeaders()
	(*headers)["x-custom"] = "value1"
	data = []byte("X-Custom: value2\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "value1, value2", (*headers)["x-custom"])
	assert.False(t, done)
}

func TestCaseInsensitiveHeaders(t *testing.T) {
	// Test: Header name in uppercase
	headers := NewHeaders()
	data := []byte("CONTENT-TYPE: application/json\r\n\r\n")
	_, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "application/json", (*headers)["content-type"])
	assert.False(t, done)

	// Test: Header name in mixed case
	headers = NewHeaders()
	data = []byte("CoNtEnT-TyPe: text/xml\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "text/xml", (*headers)["content-type"])
	assert.False(t, done)

	// Test: Header name in lowercase
	headers = NewHeaders()
	data = []byte("content-type: text/plain\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "text/plain", (*headers)["content-type"])
	assert.False(t, done)

	// Test: Two mixed case headers stored with lowercase keys
	headers = NewHeaders()
	data = []byte("HOST: example.com\r\nContent-Length: 100\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "example.com", (*headers)["host"])
	assert.Equal(t, "100", (*headers)["content-length"])
	// Verify keys are lowercase
	_, hasHost := (*headers)["HOST"]
	assert.False(t, hasHost, "uppercase key should not exist")
	_, hasLowerHost := (*headers)["host"]
	assert.True(t, hasLowerHost, "lowercase key should exist")
	assert.False(t, done)

	// Test: Accessing header by different case returns empty (only lowercase key exists)
	headers = NewHeaders()
	data = []byte("Accept-Encoding: gzip\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "gzip", (*headers)["accept-encoding"])
	assert.Equal(t, "", (*headers)["Accept-Encoding"]) // Mixed case key doesn't exist
	assert.Equal(t, "", (*headers)["ACCEPT-ENCODING"]) // Uppercase key doesn't exist
	assert.False(t, done)
}

func TestMissingEndOfHeaders(t *testing.T) {
	// Test: Single header without CRLF terminator
	headers := NewHeaders()
	data := []byte("Host: example.com")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header with only CR (no LF)
	headers = NewHeaders()
	data = []byte("Host: example.com\r")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header with only LF (no CR)
	headers = NewHeaders()
	data = []byte("Host: example.com\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Multiple headers but missing final CRLF
	headers = NewHeaders()
	data = []byte("Host: example.com\r\nContent-Type: text/html")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	// Should parse first header but not the second (incomplete)
	assert.Equal(t, "example.com", (*headers)["host"])
	assert.Equal(t, 19, n)
	assert.False(t, done)

	// Test: Headers with single CRLF at end (no double CRLF to signal end)
	headers = NewHeaders()
	data = []byte("Host: example.com\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "example.com", (*headers)["host"])
	assert.Equal(t, 19, n)
	assert.False(t, done)

	// Test: Partial CRLF sequence at end
	headers = NewHeaders()
	data = []byte("Host: example.com\r\n\r")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "example.com", (*headers)["host"])
	assert.Equal(t, 19, n)
	assert.False(t, done)

	// Test: Empty data
	headers = NewHeaders()
	data = []byte("")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
