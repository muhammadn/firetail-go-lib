package firetail

import (
	"net/http"

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

	// DisableValidation is an optional flag which, if set to true, disables request & response validation; validation is enabled by default
	DisableValidation *bool

	// FiretailEndpoint is the Firetail logging endpoint request data should be sent to
	FiretailEndpoint string
}
