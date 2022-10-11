package logging

import (
	"crypto"
	"fmt"
)

type HeaderMask int

const (
	// If the mask is being applied strictly, then the header will be removed, otherwise it will be preserved
	UnsetHeader HeaderMask = iota

	// The header will be preserved and logged to FireTail as received
	PreserveHeader

	// The header will be removed entirely and not reported to Firetail
	RemoveHeader

	// The header value will be removed and not reported to FireTail
	RemoveHeaderValues

	// The header's value will hashed before it is reported to Firetail
	HashHeaderValues

	// Both the header's name and value will be hashed before being reported to Firetail
	HashHeader
)

func MaskHeaders(unmaskedHeaders map[string][]string, headersMask map[string]HeaderMask, isStrict bool) map[string][]string {
	maskedHeaders := map[string][]string{}

	headerNameHashFails := 0
	for headerName, headerValues := range unmaskedHeaders {
		mask := headersMask[headerName]

		switch mask {
		case UnsetHeader:
			// If the mask is being applied strictly, and the headersMask is Unset for this header, we skip it
			if isStrict {
				break
			}
			// Else, we treat it as if it's preserved
			maskedHeaders[headerName] = headerValues
			break

		case PreserveHeader:
			maskedHeaders[headerName] = headerValues
			break

		case RemoveHeader:
			// Nothing to do here!
			break

		case RemoveHeaderValues:
			maskedHeaders[headerName] = []string{}
			break

		case HashHeaderValues:
			maskedHeaders[headerName] = hashValues(headerValues)
			break

		case HashHeader:
			hashedHeaderName, err := hashString(headerName)
			if err != nil {
				hashedHeaderName = fmt.Sprintf("Failed to hash header name (%d)", headerNameHashFails)
				headerNameHashFails += 1
			}
			maskedHeaders[hashedHeaderName] = hashValues(headerValues)
			break
		}
	}

	return maskedHeaders
}

func hashValues(values []string) []string {
	hashedValues := []string{}
	for _, value := range values {
		hashedValue, err := hashString(value)
		if err != nil {
			hashedValues = append(hashedValues, "Failed to hash, err: "+err.Error())
			continue
		}
		hashedValues = append(hashedValues, hashedValue)
	}
	return hashedValues
}

func hashString(value string) (string, error) {
	hasher := crypto.SHA1.New()
	_, err := hasher.Write([]byte(value))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", (hasher.Sum(nil))), nil
}
