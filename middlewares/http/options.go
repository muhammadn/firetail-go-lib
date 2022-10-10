package firetail

import (
	"net/http"
	"strings"

	"github.com/FireTail-io/firetail-go-lib/headers"
	"github.com/FireTail-io/firetail-go-lib/utils"
)

// Options is an options struct used when creating a Firetail middleware (GetMiddleware)
type Options struct {
	// SpecPath is the path at which your openapi spec can be found
	SpecPath string

	// SourceIPCallback is an optional callback func which takes the http.Request and returns the source IP of the request as a string,
	// allowing you to, for example, handle cases where your service is running behind a proxy and needs to extract the source IP from
	// the request headers instead
	SourceIPCallback func(*http.Request) string

	// ErrHandler is an optional callback func which is given an error and a ResponseWriter to which an apropriate response can be written
	// for the error. This allows you customise the responses given, when for example a request or response fails to validate against the
	// openapi spec, to be consistent with the format in which the rest of your application returns error responses
	ErrHandler func(error, *utils.ResponseWriter)

	// DisableValidation is an optional flag which, if set to true, disables request & response validation
	DisableValidation bool

	// RequestSanitisationCallback is an optional callback which is given the request body as bytes & returns a stringified request body which
	// is then logged to Firetail. This is useful for writing custom logic to redact any sensitive data from your request bodies before it is logged
	// in Firetail.
	RequestSanitisationCallback func([]byte) string

	// RequestHeadersMask is a map of header names to HeaderMask values, which can be used to control the request headers reported to Firetail
	RequestHeadersMask *map[string]headers.HeaderMask

	// RequestHeadersMaskStrict is an optional flag which, if set to true, will configure the Firetail middleware to only report request headers explicitly described in the RequestHeadersMask
	RequestHeadersMaskStrict bool

	// ResponseSanitisationCallback is an optional callback which is given the response body as bytes & returns a stringified response body which
	// is then logged to Firetail. This is useful for writing custom logic to redact any sensitive data from your response bodies before it is logged
	// in Firetail.
	ResponseSanitisationCallback func([]byte) string

	// ResponseHeadersMask is a map of header names to HeaderMask values, which can be used to control the response headers reported to Firetail
	ResponseHeadersMask *map[string]headers.HeaderMask

	// ResponseHeadersMaskStrict is an optional flag which, if set to true, will configure the Firetail middleware to only report response headers explicitly described in the ResponseHeadersMask
	ResponseHeadersMaskStrict bool

	// FiretailEndpoint is the Firetail logging endpoint request data should be sent to
	FiretailEndpoint string
}

func (o *Options) setDefaults() {
	if o.SourceIPCallback == nil {
		o.SourceIPCallback = func(r *http.Request) string {
			return strings.Split(r.RemoteAddr, ":")[0]
		}
	}

	if o.ErrHandler == nil {
		o.ErrHandler = func(err error, w *utils.ResponseWriter) {
			switch err {
			case utils.ErrPathNotFound:
				w.WriteHeader(404)
				w.Write([]byte("404 - Not Found"))
				break
			case utils.ErrMethodNotAllowed:
				w.WriteHeader(405)
				w.Write([]byte("405 - Method Not Allowed"))
				break
			case utils.ErrRequestValidationFailed:
				w.WriteHeader(400)
				w.Write([]byte("400 - Bad Request"))
				break
			case utils.ErrResponseValidationFailed:
				w.WriteHeader(500)
				w.Write([]byte("500 - Internal Server Error"))
				return
			default:
				// Even if the err is nil, we return a 500, as defaultErrHandler should never be called with a nil err
				w.WriteHeader(500)
				w.Write([]byte("500 - Internal Server Error"))
				break
			}
		}
	}

	if o.RequestSanitisationCallback == nil {
		o.RequestSanitisationCallback = func(b []byte) string {
			return string(b)
		}
	}

	if o.RequestHeadersMask == nil {
		// TODO: create default
		o.RequestHeadersMask = &map[string]headers.HeaderMask{}
	}

	if o.ResponseSanitisationCallback == nil {
		o.ResponseSanitisationCallback = func(b []byte) string {
			return string(b)
		}
	}

	if o.ResponseHeadersMask == nil {
		// TODO: create default
		o.ResponseHeadersMask = &map[string]headers.HeaderMask{}
	}
}
