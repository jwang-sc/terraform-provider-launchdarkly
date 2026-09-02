package launchdarkly

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression coverage for "Provider produced inconsistent result after
// apply" on launchdarkly_segment's unbounded_context_kind/view_keys:
// segmentRead used to unconditionally d.Set these fields from whatever the
// API returned (nil *string / empty []string), and d.Set's reflection-based
// coercion turns a nil/empty value into a concrete "" / empty-set cty value
// instead of leaving the attribute null -- tripping Terraform's
// post-apply/post-import consistency check for any segment with no real
// value for these fields (the common case: standard, non-Big segments with
// no view associations). The fix: skip the Set call entirely when there's
// no real value, so the attribute stays whatever it already was (null on a
// fresh read/import).

func testSegmentSchema() map[string]*schema.Schema {
	return baseSegmentSchema(segmentSchemaOptions{isDataSource: false})
}

func TestSetUnboundedContextKindIfPresent(t *testing.T) {
	t.Run("nil value leaves a prior value untouched (does not overwrite with empty string)", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, testSegmentSchema(), map[string]interface{}{
			UNBOUNDED_CONTEXT_KIND: "user",
		})

		err := setUnboundedContextKindIfPresent(d, nil)
		require.NoError(t, err)

		assert.Equal(t, "user", d.Get(UNBOUNDED_CONTEXT_KIND))
	})

	t.Run("real value is set", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, testSegmentSchema(), map[string]interface{}{})
		val := "user"

		err := setUnboundedContextKindIfPresent(d, &val)
		require.NoError(t, err)

		assert.Equal(t, "user", d.Get(UNBOUNDED_CONTEXT_KIND))
	})
}

func TestSetViewKeysIfNonEmpty(t *testing.T) {
	t.Run("empty slice leaves a prior value untouched (does not overwrite with an empty set)", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, testSegmentSchema(), map[string]interface{}{
			VIEW_KEYS: []interface{}{"view-1"},
		})

		err := setViewKeysIfNonEmpty(d, []string{})
		require.NoError(t, err)

		got := d.Get(VIEW_KEYS).(*schema.Set).List()
		assert.ElementsMatch(t, []interface{}{"view-1"}, got)
	})

	t.Run("nil slice leaves a prior value untouched", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, testSegmentSchema(), map[string]interface{}{
			VIEW_KEYS: []interface{}{"view-1"},
		})

		err := setViewKeysIfNonEmpty(d, nil)
		require.NoError(t, err)

		got := d.Get(VIEW_KEYS).(*schema.Set).List()
		assert.ElementsMatch(t, []interface{}{"view-1"}, got)
	})

	t.Run("non-empty slice is set", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, testSegmentSchema(), map[string]interface{}{})

		err := setViewKeysIfNonEmpty(d, []string{"view-1", "view-2"})
		require.NoError(t, err)

		got := d.Get(VIEW_KEYS).(*schema.Set).List()
		assert.ElementsMatch(t, []interface{}{"view-1", "view-2"}, got)
	})
}
