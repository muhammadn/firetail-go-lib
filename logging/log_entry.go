// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    logEntry, err := UnmarshalLogEntry(bytes)
//    bytes, err = logEntry.Marshal()

package logging

import "encoding/json"

func UnmarshalLogEntry(data []byte) (LogEntry, error) {
	var r LogEntry
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *LogEntry) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// All the information required to make a logging entry in Firetail
type LogEntry struct {
	DateCreated   int64    `json:"dateCreated"`   // The time the request was logged in UNIX milliseconds
	ExecutionTime float64  `json:"executionTime"` // The time elapsed during the execution required to respond to the request, in milliseconds
	Request       Request  `json:"request"`
	Response      Response `json:"response"`
	Version       Version  `json:"version"` // The version of the firetail logging schema used
}

type Request struct {
	Body         string              `json:"body"`         // The request body, stringified
	Headers      map[string][]string `json:"headers"`      // The request headers
	HTTPProtocol HTTPProtocol        `json:"httpProtocol"` // The HTTP protocol used in the request
	IP           string              `json:"ip"`           // The source IP of the request
	Method       Method              `json:"method"`       // The request method. Src for allowed values can be found here: <a; href='https://www.iana.org/assignments/http-methods/http-methods.xhtml#methods'>https://www.iana.org/assignments/http-methods/http-methods.xhtml#methods</a>.
	URI          string              `json:"uri"`          // The URI the request was made to
}

type Response struct {
	Body       string              `json:"body"`    // The response body, stringified
	Headers    map[string][]string `json:"headers"` // The response headers
	StatusCode int64               `json:"statusCode"`
}

// The HTTP protocol used in the request
type HTTPProtocol string

const (
	HTTP10 HTTPProtocol = "HTTP/1.0"
	HTTP11 HTTPProtocol = "HTTP/1.1"
	HTTP2  HTTPProtocol = "HTTP/2"
	HTTP3  HTTPProtocol = "HTTP/3"
)

// The request method. Src for allowed values can be found here: <a
// href='https://www.iana.org/assignments/http-methods/http-methods.xhtml#methods'>https://www.iana.org/assignments/http-methods/http-methods.xhtml#methods</a>.
type Method string

const (
	ACL               Method = "ACL"
	BaselineControl   Method = "BASELINE-CONTROL"
	Bind              Method = "BIND"
	Checkin           Method = "CHECKIN"
	Checkout          Method = "CHECKOUT"
	Connect           Method = "CONNECT"
	Copy              Method = "COPY"
	Delete            Method = "DELETE"
	Empty             Method = "*"
	Get               Method = "GET"
	Head              Method = "HEAD"
	Label             Method = "LABEL"
	Link              Method = "LINK"
	Lock              Method = "LOCK"
	Merge             Method = "MERGE"
	Mkactivity        Method = "MKACTIVITY"
	Mkcalendar        Method = "MKCALENDAR"
	Mkcol             Method = "MKCOL"
	Mkredirectref     Method = "MKREDIRECTREF"
	Mkworkspace       Method = "MKWORKSPACE"
	Move              Method = "MOVE"
	Options           Method = "OPTIONS"
	Orderpatch        Method = "ORDERPATCH"
	Patch             Method = "PATCH"
	Post              Method = "POST"
	Pri               Method = "PRI"
	Propfind          Method = "PROPFIND"
	Proppatch         Method = "PROPPATCH"
	Put               Method = "PUT"
	Rebind            Method = "REBIND"
	Report            Method = "REPORT"
	Search            Method = "SEARCH"
	Trace             Method = "TRACE"
	Unbind            Method = "UNBIND"
	Uncheckout        Method = "UNCHECKOUT"
	Unlink            Method = "UNLINK"
	Unlock            Method = "UNLOCK"
	Update            Method = "UPDATE"
	Updateredirectref Method = "UPDATEREDIRECTREF"
	VersionControl    Method = "VERSION-CONTROL"
)

// The version of the firetail logging schema used
type Version string

const (
	The100Alpha Version = "1.0.0-alpha"
)
