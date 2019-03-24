package shorty

import (
	"log"
	"net/http"
	"time"
)

// Logger common http-request logger, shows the duration of each action with a log entry before and
// after the actual handler function.
func Logger(inner http.HandlerFunc, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[%s] %s %s", name, r.Method, r.RequestURI)
		inner.ServeHTTP(w, r)
		log.Printf("[%s] %s %s (%s)", name, r.Method, r.RequestURI, time.Since(start))
	})
}
