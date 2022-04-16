package middleware

import (
	"bytes"
	"io"
	"net/http"
	"sync"

	"github.com/newtstat/cloudacme/contexts"
	"github.com/rec-logger/rec.go"
	"golang.org/x/xerrors"
)

type requestBodyBuffer struct {
	Buffer *bytes.Buffer
}

// nolint: gochecknoglobals
var requestBodyBufferPool = &sync.Pool{
	New: func() interface{} {
		return &requestBodyBuffer{
			Buffer: bytes.NewBuffer(nil),
		}
	},
}

func RequestBodyBufferMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := rec.ContextLogger(r.Context())

			// nolint: forcetypeassert
			buf := requestBodyBufferPool.Get().(*requestBodyBuffer)

			if _, err := buf.Buffer.ReadFrom(r.Body); err != nil {
				l.E().Error(xerrors.Errorf("buf.Buffer.ReadFrom: %w", err))
			}

			r.Body = io.NopCloser(buf.Buffer)

			ctx := contexts.WithRequestBodyBuffer(r.Context(), buf.Buffer)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
