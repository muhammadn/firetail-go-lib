package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrictUnsetIsRemoved(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"Test-Header": UnsetHeader,
	}
	unmaskedHeaders := map[string][]string{
		"Test-Header": {"test-value"},
	}

	maskedHeaders := Mask(unmaskedHeaders, headersMask, true)

	require.Empty(t, maskedHeaders)
}

func TestNonStrictUnsetIsPreserved(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"Test-Header": UnsetHeader,
	}
	unmaskedHeaders := map[string][]string{
		"Test-Header": {"test-value"},
	}

	maskedHeaders := Mask(unmaskedHeaders, headersMask, false)

	require.Equal(t, unmaskedHeaders, maskedHeaders)
}

func TestPreserveHeaderIsPreserved(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"Test-Header": PreserveHeader,
	}
	unmaskedHeaders := map[string][]string{
		"Test-Header": {"test-value"},
	}

	maskedHeaders := Mask(unmaskedHeaders, headersMask, true)

	require.Equal(t, unmaskedHeaders, maskedHeaders)
}

func TestRemoveHeaderIsRemoved(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"Test-Header": RemoveHeader,
	}
	unmaskedHeaders := map[string][]string{
		"Test-Header": {"test-value"},
	}

	maskedHeaders := Mask(unmaskedHeaders, headersMask, false)

	assert.Empty(t, maskedHeaders)
}

func TestRemoveHeaderValuesRemovesValues(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"Test-Header": RemoveHeaderValues,
	}
	unmaskedHeaders := map[string][]string{
		"Test-Header": {"test-value"},
	}

	maskedHeaders := Mask(unmaskedHeaders, headersMask, false)

	assert.Empty(t, maskedHeaders["Test-Header"])
}

func TestHashHeaderValuesHashesValues(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"Test-Header": HashHeaderValues,
	}
	unmaskedHeaders := map[string][]string{
		"Test-Header": {"test-value"},
	}

	maskedHeaders := Mask(unmaskedHeaders, headersMask, false)

	require.Contains(t, maskedHeaders, "Test-Header")
	assert.Len(t, maskedHeaders["Test-Header"], 1)
	assert.Contains(
		t,
		maskedHeaders["Test-Header"],
		"1382103331d56fa62a3f0b12388aad5cdb36389d",
	)
}

func TestHashHeaderValuesHashesMultipleValues(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"Test-Header": HashHeaderValues,
	}
	unmaskedHeaders := map[string][]string{
		"Test-Header": {"test-value-1", "test-value-2"},
	}

	maskedHeaders := Mask(unmaskedHeaders, headersMask, false)

	require.Contains(t, maskedHeaders, "Test-Header")
	assert.Len(t, maskedHeaders["Test-Header"], 2)
	assert.Contains(
		t,
		maskedHeaders["Test-Header"],
		"30ecdd733caf1b69c94a4c905a419bf4053876e4",
	)
	assert.Contains(
		t,
		maskedHeaders["Test-Header"],
		"5b75855d18cf0ce802922549e002373be4d0f6de",
	)
}

func TestHashHeaderHashesHeader(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"Test-Header": HashHeader,
	}
	unmaskedHeaders := map[string][]string{
		"Test-Header": {"test-value"},
	}

	maskedHeaders := Mask(unmaskedHeaders, headersMask, false)

	require.Contains(t, maskedHeaders, "e611b81f6ddc182d400d28bac79d0aac581afeb8")
	assert.Len(t, maskedHeaders["e611b81f6ddc182d400d28bac79d0aac581afeb8"], 1)
	assert.Contains(
		t,
		maskedHeaders["e611b81f6ddc182d400d28bac79d0aac581afeb8"],
		"1382103331d56fa62a3f0b12388aad5cdb36389d",
	)
}

func TestHashHeaderHashesMultipleHeaders(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"Test-Header-1": HashHeader,
		"Test-Header-2": HashHeader,
	}
	unmaskedHeaders := map[string][]string{
		"Test-Header-1": {"test-value"},
		"Test-Header-2": {"test-value"},
	}

	maskedHeaders := Mask(unmaskedHeaders, headersMask, false)

	assert.Len(t, maskedHeaders, 2)

	require.Contains(t, maskedHeaders, "43890d9e7edd04b58cd7244d78b4e08823d2875a")
	assert.Len(t, maskedHeaders["43890d9e7edd04b58cd7244d78b4e08823d2875a"], 1)
	assert.Contains(
		t, maskedHeaders["43890d9e7edd04b58cd7244d78b4e08823d2875a"],
		"1382103331d56fa62a3f0b12388aad5cdb36389d",
	)

	require.Contains(t, maskedHeaders, "e0e8917877aa1bb1c68839ff38612a7b1987647e")
	assert.Len(t, maskedHeaders["e0e8917877aa1bb1c68839ff38612a7b1987647e"], 1)
	assert.Contains(
		t, maskedHeaders["e0e8917877aa1bb1c68839ff38612a7b1987647e"],
		"1382103331d56fa62a3f0b12388aad5cdb36389d",
	)
}
