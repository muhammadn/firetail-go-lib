package logging

import (
	"crypto"
	"fmt"
	"log"
	"regexp"
	"strings"
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

	// Any of the header's values which match against a JWT pattern are removed
	RedactJWTSignature
)

func MaskHeaders(unmaskedHeaders map[string][]string, headersMask map[string]HeaderMask, isStrict bool) map[string][]string {
	maskedHeaders := map[string][]string{}

	for headerName, headerValues := range unmaskedHeaders {
		mask := headersMask[strings.ToLower(headerName)]

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
			hashedHeaderName := hashString(headerName)
			maskedHeaders[hashedHeaderName] = hashValues(headerValues)
			break

		case RedactJWTSignature:
			for i, value := range headerValues {
				isJWT, _ := regexp.MatchString(
					"[Bb]earer [A-Za-z0-9-_]*.[A-Za-z0-9-_]*.[A-Za-z0-9-_]*",
					value,
				)
				log.Println("isJWT", isJWT, headerName, value)
				if !isJWT {
					continue
				}
				headerValues[i] = strings.Join(
					// Indexing [:2] here should not fail as the regex completed earlier should guarantee that we have
					// two '.' characters, so `strings.Split(value, ".")` should return a slice containing three elements
					strings.Split(value, ".")[:2],
					".",
				)
			}
			maskedHeaders[headerName] = headerValues
			break
		}

	}

	return maskedHeaders
}

func hashValues(values []string) []string {
	hashedValues := []string{}
	for _, value := range values {
		hashedValue := hashString(value)
		hashedValues = append(hashedValues, hashedValue)
	}
	return hashedValues
}

func hashString(value string) string {
	hasher := crypto.SHA1.New()
	hasher.Write([]byte(value)) // SHA1's Write implementation never returns a non-nil err.
	return fmt.Sprintf("%x", (hasher.Sum(nil)))
}
