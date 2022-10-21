package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrictUnsetIsRemoved(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header": UnsetHeader,
	}
	unmaskedHeaders := map[string][]string{
		"test-header": {"test-value"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, true)

	require.Empty(t, maskedHeaders)
}

func TestNonStrictUnsetIsPreserved(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header": UnsetHeader,
	}
	unmaskedHeaders := map[string][]string{
		"test-header": {"test-value"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, false)

	require.Equal(t, unmaskedHeaders, maskedHeaders)
}

func TestPreserveHeaderIsPreserved(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header": PreserveHeader,
	}
	unmaskedHeaders := map[string][]string{
		"test-header": {"test-value"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, true)

	require.Equal(t, unmaskedHeaders, maskedHeaders)
}

func TestRemoveHeaderIsRemoved(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header": RemoveHeader,
	}
	unmaskedHeaders := map[string][]string{
		"test-header": {"test-value"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, false)

	assert.Empty(t, maskedHeaders)
}

func TestRemoveHeaderValuesRemovesValues(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header": RemoveHeaderValues,
	}
	unmaskedHeaders := map[string][]string{
		"test-header": {"test-value"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, false)

	assert.Empty(t, maskedHeaders["test-header"])
}

func TestHashHeaderValuesHashesValues(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header": HashHeaderValues,
	}
	unmaskedHeaders := map[string][]string{
		"test-header": {"test-value"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, false)

	require.Contains(t, maskedHeaders, "test-header")
	assert.Len(t, maskedHeaders["test-header"], 1)
	assert.Contains(
		t,
		maskedHeaders["test-header"],
		"1382103331d56fa62a3f0b12388aad5cdb36389d",
	)
}

func TestHashHeaderValuesHashesMultipleValues(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header": HashHeaderValues,
	}
	unmaskedHeaders := map[string][]string{
		"test-header": {"test-value-1", "test-value-2"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, false)

	require.Contains(t, maskedHeaders, "test-header")
	assert.Len(t, maskedHeaders["test-header"], 2)
	assert.Contains(
		t,
		maskedHeaders["test-header"],
		"30ecdd733caf1b69c94a4c905a419bf4053876e4",
	)
	assert.Contains(
		t,
		maskedHeaders["test-header"],
		"5b75855d18cf0ce802922549e002373be4d0f6de",
	)
}

func TestHashHeaderHashesHeader(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header": HashHeader,
	}
	unmaskedHeaders := map[string][]string{
		"test-header": {"test-value"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, false)

	require.Contains(t, maskedHeaders, "ae2ad4d388518f06280f224718538ab4b6d4ba13")
	assert.Len(t, maskedHeaders["ae2ad4d388518f06280f224718538ab4b6d4ba13"], 1)
	assert.Contains(
		t,
		maskedHeaders["ae2ad4d388518f06280f224718538ab4b6d4ba13"],
		"1382103331d56fa62a3f0b12388aad5cdb36389d",
	)
}

func TestHashHeaderHashesMultipleHeaders(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header-1": HashHeader,
		"test-header-2": HashHeader,
	}
	unmaskedHeaders := map[string][]string{
		"test-header-1": {"test-value"},
		"test-header-2": {"test-value"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, false)

	assert.Len(t, maskedHeaders, 2)

	require.Contains(t, maskedHeaders, "c50b5cf22d879f3190720f425bd5724e554d1be5")
	assert.Len(t, maskedHeaders["c50b5cf22d879f3190720f425bd5724e554d1be5"], 1)
	assert.Contains(
		t, maskedHeaders["c50b5cf22d879f3190720f425bd5724e554d1be5"],
		"1382103331d56fa62a3f0b12388aad5cdb36389d",
	)

	require.Contains(t, maskedHeaders, "c65eb1fbd6f25c15e625cb08e356dfb04227663f")
	assert.Len(t, maskedHeaders["c65eb1fbd6f25c15e625cb08e356dfb04227663f"], 1)
	assert.Contains(
		t, maskedHeaders["c65eb1fbd6f25c15e625cb08e356dfb04227663f"],
		"1382103331d56fa62a3f0b12388aad5cdb36389d",
	)
}

func TestRedactJWTSignatureRedactsJWTSignature(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header": RedactJWTSignature,
	}
	unmaskedHeaders := map[string][]string{
		"test-header": {"bearer header.payload.signature"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, false)

	require.Contains(t, maskedHeaders, "test-header")
	assert.Len(t, maskedHeaders["test-header"], 1)
	assert.Contains(
		t,
		maskedHeaders["test-header"],
		"bearer header.payload",
	)
}

func TestRedactJWTSignatureRetainsNonJWTValue(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header": RedactJWTSignature,
	}
	unmaskedHeaders := map[string][]string{
		"test-header": {"test-value"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, false)

	require.Contains(t, maskedHeaders, "test-header")
	assert.Len(t, maskedHeaders["test-header"], 1)
	assert.Contains(
		t,
		maskedHeaders["test-header"],
		"test-value",
	)
}

func TestRedactJWTSignatureHandlesMultipleValues(t *testing.T) {
	headersMask := map[string]HeaderMask{
		"test-header-1": RedactJWTSignature,
		"test-header-2": RedactJWTSignature,
		"test-header-3": RedactJWTSignature,
	}
	unmaskedHeaders := map[string][]string{
		"test-header-1": {"test-value", "bearer header.payload.signature"},
		"test-header-2": {"test-value"},
		"test-header-3": {"bearer header.payload.signature"},
	}

	maskedHeaders := MaskHeaders(unmaskedHeaders, headersMask, false)

	require.Contains(t, maskedHeaders, "test-header-1")
	assert.Len(t, maskedHeaders["test-header-1"], 2)
	assert.Contains(t, maskedHeaders["test-header-1"], "test-value")
	assert.Contains(t, maskedHeaders["test-header-1"], "bearer header.payload")

	require.Contains(t, maskedHeaders, "test-header-2")
	assert.Len(t, maskedHeaders["test-header-2"], 1)
	assert.Contains(t, maskedHeaders["test-header-2"], "test-value")

	require.Contains(t, maskedHeaders, "test-header-3")
	assert.Len(t, maskedHeaders["test-header-3"], 1)
	assert.Contains(t, maskedHeaders["test-header-3"], "bearer header.payload")
}
