package pal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhulik/pal"
)

// TestParseTag tests the ParseTag function
func TestParseTag(t *testing.T) {
	t.Parallel()

	t.Run("parses single tag without value", func(t *testing.T) {
		t.Parallel()

		tags, err := pal.ParseTag("skip")

		assert.NoError(t, err)
		assert.Equal(t, map[pal.Tag]string{
			pal.TagSkip: "",
		}, tags)
	})

	t.Run("parses single tag with value", func(t *testing.T) {
		t.Parallel()

		tags, err := pal.ParseTag("name=MyService")

		assert.NoError(t, err)
		assert.Equal(t, map[pal.Tag]string{
			pal.TagName: "MyService",
		}, tags)
	})

	t.Run("parses multiple tags without values", func(t *testing.T) {
		t.Parallel()

		tags, err := pal.ParseTag("skip,match_interface")

		assert.NoError(t, err)
		assert.Equal(t, map[pal.Tag]string{
			pal.TagSkip:           "",
			pal.TagMatchInterface: "",
		}, tags)
	})

	t.Run("parses multiple tags with values", func(t *testing.T) {
		t.Parallel()

		tags, err := pal.ParseTag("name=MyService,match_interface=MyInterface")

		assert.NoError(t, err)
		assert.Equal(t, map[pal.Tag]string{
			pal.TagName:           "MyService",
			pal.TagMatchInterface: "MyInterface",
		}, tags)
	})

	t.Run("parses mixed tags with and without values", func(t *testing.T) {
		t.Parallel()

		tags, err := pal.ParseTag("skip,name=MyService,match_interface")

		assert.NoError(t, err)
		assert.Equal(t, map[pal.Tag]string{
			pal.TagSkip:           "",
			pal.TagName:           "MyService",
			pal.TagMatchInterface: "",
		}, tags)
	})

	t.Run("parses tag with empty value", func(t *testing.T) {
		t.Parallel()

		tags, err := pal.ParseTag("name=")

		assert.NoError(t, err)
		assert.Equal(t, map[pal.Tag]string{
			pal.TagName: "",
		}, tags)
	})

	t.Run("handles empty input as unsupported tag", func(t *testing.T) {
		t.Parallel()

		tags, err := pal.ParseTag("")

		assert.NoError(t, err)
		assert.Empty(t, tags)
	})

	t.Run("handles whitespace around tags as unsupported tags", func(t *testing.T) {
		t.Parallel()

		tags, err := pal.ParseTag(" skip , name=MyService ")

		assert.NoError(t, err)
		assert.Equal(t, map[pal.Tag]string{
			pal.TagSkip: "",
			pal.TagName: "MyService",
		}, tags)
	})

	t.Run("returns error for unsupported tag", func(t *testing.T) {
		t.Parallel()

		_, err := pal.ParseTag("unsupported")

		assert.Error(t, err)
		assert.ErrorIs(t, err, pal.ErrInvalidTag)
		assert.Contains(t, err.Error(), "tag unsupported unsupported")
	})

	t.Run("returns error for unsupported tag in multiple tags", func(t *testing.T) {
		t.Parallel()

		_, err := pal.ParseTag("skip,unsupported,name=MyService")

		assert.Error(t, err)
		assert.ErrorIs(t, err, pal.ErrInvalidTag)
		assert.Contains(t, err.Error(), "tag unsupported unsupported")
	})

	t.Run("handles tag with multiple equals signs correctly", func(t *testing.T) {
		t.Parallel()

		_, err := pal.ParseTag("name=key=value=extra")

		assert.ErrorIs(t, err, pal.ErrInvalidTag)
		assert.Contains(t, err.Error(), "tag is malformed name=key=value=extra")
	})

	t.Run("returns error for tag with only equals", func(t *testing.T) {
		t.Parallel()

		_, err := pal.ParseTag("=")

		assert.Error(t, err)
		assert.ErrorIs(t, err, pal.ErrInvalidTag)
		assert.Contains(t, err.Error(), "tag unsupported ")
	})

	t.Run("handles tag with multiple equals and no value correctly", func(t *testing.T) {
		t.Parallel()

		_, err := pal.ParseTag("name==")

		assert.ErrorIs(t, err, pal.ErrInvalidTag)
	})

	t.Run("handles all supported tags", func(t *testing.T) {
		t.Parallel()

		tags, err := pal.ParseTag("skip,name=TestService,match_interface=TestInterface")

		assert.NoError(t, err)
		assert.Equal(t, map[pal.Tag]string{
			pal.TagSkip:           "",
			pal.TagName:           "TestService",
			pal.TagMatchInterface: "TestInterface",
		}, tags)
	})

	t.Run("handles duplicate tags by overwriting", func(t *testing.T) {
		t.Parallel()

		tags, err := pal.ParseTag("name=FirstService,name=SecondService")

		assert.NoError(t, err)
		assert.Equal(t, map[pal.Tag]string{
			pal.TagName: "SecondService",
		}, tags)
	})
}
