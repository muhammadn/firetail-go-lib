package firetail

import (
	"net/http"
	"time"
)

func handleWithoutValidation(w http.ResponseWriter, r *http.Request, next http.Handler) time.Duration {
	// There's no validation to do; we've just got to record the execution time
	startTime := time.Now()
	next.ServeHTTP(w, r)
	return time.Since(startTime)
}
