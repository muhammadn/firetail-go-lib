package logging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshalYieldsSameResult(t *testing.T) {
	testLogEntry := LogEntry{
		DateCreated:   1666356293610,
		ExecutionTime: 425,
		Request: Request{
			Body: "this is a request body",
			Headers: map[string][]string{
				"Content-Type": {"text/plain"},
			},
			HTTPProtocol: HTTP2,
			IP:           "8.8.8.8",
			Method:       Post,
			URI:          "http://firetail.io/not-real",
			Resource:     "/not-real",
		},
		Response: Response{
			Body: "this is a response body",
			Headers: map[string][]string{
				"Content-Type": {"text/plain"},
			},
			StatusCode: 200,
		},
		Version: "1.0.0-alpha",
	}

	marshalledTestLogEntry, err := testLogEntry.Marshal()
	require.Nil(t, err)

	unmarshalledTestLogEntry, err := UnmarshalLogEntry(marshalledTestLogEntry)
	require.Nil(t, err)
	require.Equal(t, testLogEntry, unmarshalledTestLogEntry)

	remarshalledTestLogEntry, err := unmarshalledTestLogEntry.Marshal()
	require.Nil(t, err)
	require.Equal(t, marshalledTestLogEntry, remarshalledTestLogEntry)
}
