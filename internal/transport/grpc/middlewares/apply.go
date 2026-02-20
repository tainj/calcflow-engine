package middlewares

import "net/http"

type Middleware func(next http.Handler) http.Handler

func Apply(handler http.Handler, mwares ...Middleware) http.Handler {
	for _, mw := range mwares {
		handler = mw(handler)
	}
	return handler
}
