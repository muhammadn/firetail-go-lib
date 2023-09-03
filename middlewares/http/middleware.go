package firetail

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// GetMiddleware creates & returns a firetail middleware. Errs if the openapi spec can't be found, validated, or loaded into a gorillamux router.
func GetMiddleware(options *Options) (func(next http.Handler) http.Handler, error) {
	options.setDefaults() // Fill in any defaults where apropriate

	// Load in our appspec, validate it & create a router from it if we have an appspec to load
	var router routers.Router
	if options.OpenapiSpecPath != "" {
		loader := &openapi3.Loader{Context: context.Background(), IsExternalRefsAllowed: true}
		//doc, err := loader.LoadFromFile(options.OpenapiSpecPath)
		doc, err := loader.LoadFromData([]byte(options.OpenapiSpecData))
		if err != nil {
			return nil, ErrorInvalidConfiguration{err}
		}
		err = doc.Validate(context.Background())
		if err != nil {
			return nil, ErrorAppspecInvalid{err}
		}
		router, err = gorillamux.NewRouter(doc)
		if err != nil {
			return nil, err
		}
	}

	// Register any custom body decoders
	for contentType, bodyDecoder := range options.CustomBodyDecoders {
		openapi3filter.RegisterBodyDecoder(contentType, bodyDecoder)
	}

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a Firetail ResponseWriter so we can access the response body, status code etc. for logging & validation later
			localResponseWriter := httptest.NewRecorder()

			// Read in the request body so we can log it & replace r.Body with a new copy for the next http.Handler to read from
			requestBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				options.ErrCallback(ErrorAtRequestUnspecified{err}, localResponseWriter, r)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

			// Check there's a corresponding route for this request if we have a router
			var route *routers.Route
			var pathParams map[string]string
			if router != nil && (options.EnableRequestValidation || options.EnableResponseValidation) {
				route, pathParams, err = router.FindRoute(r)
				if err == routers.ErrMethodNotAllowed {
					options.ErrCallback(ErrorUnsupportedMethod{r.URL.Path, r.Method}, localResponseWriter, r)
					return
				} else if err == routers.ErrPathNotFound {
					options.ErrCallback(ErrorRouteNotFound{r.URL.Path}, localResponseWriter, r)
					return
				} else if err != nil {
					options.ErrCallback(ErrorAtRequestUnspecified{err}, localResponseWriter, r)
					return
				}
				// We now know the resource that was requested, so we can fill it into our log entry
			} 

			// If it has been enabled, and we were able to determine the route and path params, validate the request against the openapi spec
			if options.EnableRequestValidation && route != nil && pathParams != nil {
				requestValidationInput := &openapi3filter.RequestValidationInput{
					Request:    r,
					PathParams: pathParams,
					Route:      route,
					Options: &openapi3filter.Options{
						AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
							authCallback, hasAuthCallback := options.AuthCallbacks[ai.SecuritySchemeName]
							if !hasAuthCallback {
								return ErrorAuthSchemeNotImplemented{ai.SecuritySchemeName}
							}
							return authCallback(ctx, ai)
						},
					},
				}
				err = openapi3filter.ValidateRequest(context.Background(), requestValidationInput)
				if err != nil {
					// If the err is an openapi3filter RequestError, we can extract more information from the err...
					if err, isRequestErr := err.(*openapi3filter.RequestError); isRequestErr {
						// TODO: Using strings.Contains is janky here and may break - should replace with something more reliable
						// See the following open issue on the kin-openapi repo: https://github.com/getkin/kin-openapi/issues/477
						// TODO: Open source contribution to kin-openapi?
						if strings.Contains(err.Reason, "header Content-Type has unexpected value") {
							options.ErrCallback(ErrorRequestContentTypeInvalid{r.Header.Get("Content-Type"), route.Path}, localResponseWriter, r)
							return
						}
						if strings.Contains(err.Error(), "body has an error") {
							options.ErrCallback(ErrorRequestBodyInvalid{err}, localResponseWriter, r)
							return
						}
						if strings.Contains(err.Error(), "header has an error") {
							options.ErrCallback(ErrorRequestHeadersInvalid{err}, localResponseWriter, r)
							return
						}
						if strings.Contains(err.Error(), "query has an error") {
							options.ErrCallback(ErrorRequestQueryParamsInvalid{err}, localResponseWriter, r)
							return
						}
						if strings.Contains(err.Error(), "path has an error") {
							options.ErrCallback(ErrorRequestPathParamsInvalid{err}, localResponseWriter, r)
							return
						}
					}

					// If the validation fails due to a security requirement, we pass a SecurityRequirementsError to the ErrCallback
					if err, isSecurityErr := err.(*openapi3filter.SecurityRequirementsError); isSecurityErr {
						options.ErrCallback(ErrorAuthNoMatchingScheme{err}, localResponseWriter, r)
						return
					}

					// Else, we just use a non-specific ValidationError error
					options.ErrCallback(ErrorAtRequestUnspecified{err}, localResponseWriter, r)
					return
				}
			}

			// Serve the next handler down the chain & take note of the execution time
			chainResponseWriter := httptest.NewRecorder()
			next.ServeHTTP(chainResponseWriter, r)

			// If it has been enabled, and we were able to determine the route and path params, validate the response against the openapi spec
			if options.EnableResponseValidation && route != nil && pathParams != nil {
				responseValidationInput := &openapi3filter.ResponseValidationInput{
					RequestValidationInput: &openapi3filter.RequestValidationInput{
						Request:    r,
						PathParams: pathParams,
						Route:      route,
					},
					Status: chainResponseWriter.Result().StatusCode,
					Header: chainResponseWriter.Header(),
					Options: &openapi3filter.Options{
						IncludeResponseStatus: true,
					},
				}
				responseBytes, err := ioutil.ReadAll(chainResponseWriter.Result().Body)
				if err != nil {
					options.ErrCallback(ErrorResponseBodyInvalid{err}, localResponseWriter, r)
					return
				}
				responseValidationInput.SetBodyBytes(responseBytes)
				err = openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
				if err != nil {
					if responseError, isResponseError := err.(*openapi3filter.ResponseError); isResponseError {
						if responseError.Reason == "response body doesn't match the schema" {
							options.ErrCallback(ErrorResponseBodyInvalid{responseError}, localResponseWriter, r)
							return
						} else if responseError.Reason == "status is not supported" {
							options.ErrCallback(ErrorResponseStatusCodeInvalid{responseError.Input.Status}, localResponseWriter, r)
							return
						}
					}
					options.ErrCallback(ErrorAtRequestUnspecified{err}, localResponseWriter, r)
					return
				}
			}

			// If the response written down the chain passed all of the enabled validation, we can now write it to our localResponseWriter
			for key, vals := range chainResponseWriter.HeaderMap {
				for _, val := range vals {
					localResponseWriter.Header().Add(key, val)
				}
			}
			localResponseWriter.WriteHeader(chainResponseWriter.Code)
			localResponseWriter.Write(chainResponseWriter.Body.Bytes())
		})
	}

	return middleware, nil
}
