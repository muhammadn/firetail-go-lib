package firetail

import (
	"net/http"
	"strings"

	"github.com/FireTail-io/firetail-go-lib/utils"
)

func defaultSourceIPCallback(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}

func defaultErrHandler(err error, w *utils.ResponseWriter) {
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
		w.Write([]byte("400 - Method Not Allowed"))
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

// TODO
func getDefaultRequestHeadersMask() *map[string]utils.HeaderMask {
	return &map[string]utils.HeaderMask{}
}

// TODO
func getDefaultResponseHeadersMask() *map[string]utils.HeaderMask {
	return &map[string]utils.HeaderMask{}
}
