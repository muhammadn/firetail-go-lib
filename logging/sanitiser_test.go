package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSanitiserHashesRequestHeaders(t *testing.T) {
	sanitiser := DefaultSanitiser()
	logEntry := LogEntry{
		Request: Request{
			Headers: map[string][]string{
				"set-cookie":    {"set-cookie"},
				"cookie":        {"cookie"},
				"authorization": {"authorization"},
				"x-api-key":     {"x-api-key"},
				"token":         {"token"},
				"api-token":     {"api-token"},
				"api-key":       {"api-key"},
			},
		},
	}
	sanitisedLogEntry := sanitiser(logEntry)

	for headerName, headerValues := range sanitisedLogEntry.Request.Headers {
		assert.Len(t, headerValues, 1)
		assert.Equal(t, hashString(headerName), headerValues[0])
	}
}

func TestCustomSanitiserHashesRequestHeaders(t *testing.T) {
	sanitiser := GetSanitiser(SanitiserOptions{
		RequestHeadersMask: map[string]HeaderMask{
			"set-cookie":    HashHeaderValues,
			"cookie":        HashHeaderValues,
			"authorization": HashHeaderValues,
			"x-api-key":     HashHeaderValues,
			"token":         HashHeaderValues,
			"api-token":     HashHeaderValues,
			"api-key":       HashHeaderValues,
		},
	})
	logEntry := LogEntry{
		Request: Request{
			Headers: map[string][]string{
				"set-cookie":    {"set-cookie"},
				"cookie":        {"cookie"},
				"authorization": {"authorization"},
				"x-api-key":     {"x-api-key"},
				"token":         {"token"},
				"api-token":     {"api-token"},
				"api-key":       {"api-key"},
			},
		},
	}
	sanitisedLogEntry := sanitiser(logEntry)

	for headerName, headerValues := range sanitisedLogEntry.Request.Headers {
		assert.Len(t, headerValues, 1)
		assert.Equal(t, hashString(headerName), headerValues[0])
	}
}
