package firetail

import (
	"net/http"

	"github.com/FireTail-io/firetail-go-lib/logging"
	"github.com/getkin/kin-openapi/openapi3filter"
)

// Options is an options struct used when creating a Firetail middleware (GetMiddleware)
type Options struct {
	// SpecPath is the path at which your openapi spec can be found
	OpenapiSpecPath string

	// LogApiKey is the API key which will be used when sending logs to the Firetail logging API. This value should typically be loaded
	// in from an environment variable.
	LogApiKey string

	// LogBatchCallback is an optional callback which is provided with a batch of Firetail log entries ready to be sent to Firetail. The
	// default callback sends log entries to the Firetail logging API. It may be customised to, for example, additionally log the entries
	// to a file on disk. If it returns a non-nil error, the batch will be retried later.
	// TODO
	// LogBatchCallback func([][]byte) error

	// ErrCallback is an optional callback func which is given an error and a ResponseWriter to which an apropriate response can be written
	// for the error. This allows you customise the responses given, when for example a request or response fails to validate against the
	// openapi spec, to be consistent with the format in which the rest of your application returns error responses
	ErrCallback func(error, http.ResponseWriter)

	// AuthenticationFunc is a callback func which must be defined if you wish to use security schemas in your openapi specification. See
	// the openapi3filter package's reference for further documentation, and the Chi example for a demonstration of various auth types in use:
	// https://github.com/FireTail-io/firetail-go-lib/tree/main/examples/chi
	AuthCallback openapi3filter.AuthenticationFunc

	// DisableValidation is an optional flag which, if set to true, disables request & response validation
	DisableValidation bool

	// CustomBodyDecoders is a map of Content-Type header values to openapi3 decoders - if the kin-openapi module does not support your
	// Content-Type by default, you will need to add a custom decoder here.
	CustomBodyDecoders map[string]openapi3filter.BodyDecoder

	// LogEntrySanitiser is a function used to sanitise the log entries sent to Firetail. You may wish to use this to redact sensitive
	// information, or anonymise identifiable information using a custom implementation of this callback for your application. A default
	// implementation is provided in the firetail logging package.
	LogEntrySanitiser func(logging.LogEntry) logging.LogEntry
}

func (o *Options) setDefaults() {
	if o.ErrCallback == nil {
		o.ErrCallback = func(err error, w http.ResponseWriter) {
			w.Header().Add("Content-Type", "text/plain")
			if validationErr, isValidationErr := err.(*ValidationError); isValidationErr {
				switch validationErr.Target {
				case Request:
					w.WriteHeader(400)
					w.Write([]byte("400 (Bad Request): " + err.Error()))
				case Response:
					w.WriteHeader(500)
					w.Write([]byte("500 (Internal Server Error): " + err.Error()))
				}
			} else if _, isSecurityRequirementsErr := err.(*SecurityRequirementsError); isSecurityRequirementsErr {
				w.WriteHeader(401)
				w.Header().Add("WWW-Authenticate", "Basic realm=\"User Visible Realm\"")
				w.Write([]byte("401 (Unauthorized): " + err.Error()))
			} else if _, isPathNotFoundErr := err.(*RouteNotFoundError); isPathNotFoundErr {
				w.WriteHeader(404)
				w.Write([]byte("404 (Not Found): " + err.Error()))
			} else if _, isMethodNotAllowedErr := err.(*MethodNotAllowedError); isMethodNotAllowedErr {
				w.WriteHeader(405)
				w.Write([]byte("405 (Method Not Allowed): " + err.Error()))
			} else if _, isContentTypeErr := err.(*ContentTypeError); isContentTypeErr {
				w.WriteHeader(415)
				w.Write([]byte("415 (Unsupported Media Type): " + err.Error()))
			} else {
				// Even if the err is nil, we return a 500, as defaultErrCallback should never be called with a nil err
				w.WriteHeader(500)
				w.Write([]byte("500 (Internal Server Error): " + err.Error()))
			}
		}
	}

	if o.LogEntrySanitiser == nil {
		o.LogEntrySanitiser = logging.DefaultSanitiser()
	}
}
