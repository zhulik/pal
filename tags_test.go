package pal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseTag tests the ParseTag function
func TestParseTag(t *testing.T) {
	t.Parallel()

	t.Run("parses single tag without value", func(t *testing.T) {
		t.Parallel()

		tags, err := parseTag("skip")

		assert.NoError(t, err)
		assert.Equal(t, map[Tag]string{
			TagSkip: "",
		}, tags)
	})

	t.Run("parses single tag with value", func(t *testing.T) {
		t.Parallel()

		tags, err := parseTag("name=MyService")

		assert.NoError(t, err)
		assert.Equal(t, map[Tag]string{
			TagName: "MyService",
		}, tags)
	})

	t.Run("parses multiple tags without values", func(t *testing.T) {
		t.Parallel()

		tags, err := parseTag("skip,match_interface")

		assert.NoError(t, err)
		assert.Equal(t, map[Tag]string{
			TagSkip:           "",
			TagMatchInterface: "",
		}, tags)
	})

	t.Run("parses multiple tags with values", func(t *testing.T) {
		t.Parallel()

		tags, err := parseTag("name=MyService,match_interface=MyInterface")

		assert.NoError(t, err)
		assert.Equal(t, map[Tag]string{
			TagName:           "MyService",
			TagMatchInterface: "MyInterface",
		}, tags)
	})

	t.Run("parses mixed tags with and without values", func(t *testing.T) {
		t.Parallel()

		tags, err := parseTag("skip,name=MyService,match_interface")

		assert.NoError(t, err)
		assert.Equal(t, map[Tag]string{
			TagSkip:           "",
			TagName:           "MyService",
			TagMatchInterface: "",
		}, tags)
	})

	t.Run("handles empty input as unsupported tag", func(t *testing.T) {
		t.Parallel()

		tags, err := parseTag("")

		assert.NoError(t, err)
		assert.Empty(t, tags)
	})

	t.Run("handles whitespace around tags as unsupported tags", func(t *testing.T) {
		t.Parallel()

		tags, err := parseTag(" skip , name=MyService ")

		assert.NoError(t, err)
		assert.Equal(t, map[Tag]string{
			TagSkip: "",
			TagName: "MyService",
		}, tags)
	})

	t.Run("returns error for unsupported tag", func(t *testing.T) {
		t.Parallel()

		_, err := parseTag("unsupported")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTag)
		assert.Contains(t, err.Error(), "tag unsupported unsupported")
	})

	t.Run("returns error for unsupported tag in multiple tags", func(t *testing.T) {
		t.Parallel()

		_, err := parseTag("skip,unsupported,name=MyService")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTag)
		assert.Contains(t, err.Error(), "tag unsupported unsupported")
	})

	t.Run("handles tag with multiple equals signs correctly", func(t *testing.T) {
		t.Parallel()

		_, err := parseTag("name=key=value=extra")

		assert.ErrorIs(t, err, ErrInvalidTag)
		assert.Contains(t, err.Error(), "tag is malformed name=key=value=extra")
	})

	t.Run("returns error for tag with only equals", func(t *testing.T) {
		t.Parallel()

		_, err := parseTag("=")

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTag)
		assert.Contains(t, err.Error(), "tag unsupported ")
	})

	t.Run("handles tag with multiple equals", func(t *testing.T) {
		t.Parallel()

		_, err := parseTag("name==")

		assert.ErrorIs(t, err, ErrInvalidTag)
	})

	t.Run("handles tag with no value correctly", func(t *testing.T) {
		t.Parallel()

		_, err := parseTag("name=")

		assert.ErrorIs(t, err, ErrInvalidTag)
	})

	t.Run("handles all supported tags", func(t *testing.T) {
		t.Parallel()

		tags, err := parseTag("skip,name=TestService,match_interface=TestInterface")

		assert.NoError(t, err)
		assert.Equal(t, map[Tag]string{
			TagSkip:           "",
			TagName:           "TestService",
			TagMatchInterface: "TestInterface",
		}, tags)
	})

	t.Run("handles duplicate tags by overwriting", func(t *testing.T) {
		t.Parallel()

		tags, err := parseTag("name=FirstService,name=SecondService")

		assert.NoError(t, err)
		assert.Equal(t, map[Tag]string{
			TagName: "SecondService",
		}, tags)
	})
}
