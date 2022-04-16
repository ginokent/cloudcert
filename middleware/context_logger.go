package middleware

import (
	"net/http"

	"github.com/rec-logger/rec.go"
)

func ContextLoggerRequestMiddleware(original *rec.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctxWithLogger := rec.ContextWithLogger(r.Context(),
				original.With(
					rec.String("remoteAddr", r.RemoteAddr),
					rec.String("method", r.Method),
					rec.String("host", r.Host),
					rec.String("path", r.URL.Path),
					rec.String("proto", r.Proto),
					rec.Int64("requestContentLength", r.ContentLength),
					rec.String("referer", r.Referer()),
					rec.String("userAgent", r.UserAgent()),
				),
			)

			next.ServeHTTP(rw, r.WithContext(ctxWithLogger))
		})
	}
}
