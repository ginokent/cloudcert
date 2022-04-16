package middleware

import (
	"bytes"
	"io"
	"net/http"
	"sync"

	"github.com/rec-logger/rec.go"
	"golang.org/x/xerrors"
)

type loggerResponseWriter struct {
	http.ResponseWriter
	responseBodyBuffer *responseBodyBuffer
	statusCode         int
	logger             *rec.Logger
}

type responseBodyBuffer struct {
	Buffer *bytes.Buffer
}

// nolint: gochecknoglobals
var responseBodyBufferPool = &sync.Pool{
	New: func() interface{} {
		return &responseBodyBuffer{
			Buffer: bytes.NewBuffer(nil),
		}
	},
}

func newLoggingResponseWriter(w http.ResponseWriter, l *rec.Logger) (lrw *loggerResponseWriter, put func(*responseBodyBuffer)) {
	// nolint: forcetypeassert
	buf := responseBodyBufferPool.Get().(*responseBodyBuffer)
	buf.Buffer.Reset()

	return &loggerResponseWriter{
		ResponseWriter:     w,
		responseBodyBuffer: buf,
		statusCode:         http.StatusOK,
		logger:             l,
	}, func(buf *responseBodyBuffer) { responseBodyBufferPool.Put(buf) }
}

func (lrw *loggerResponseWriter) WriteHeader(status int) {
	lrw.statusCode = status
	lrw.ResponseWriter.WriteHeader(status)
}

func (lrw *loggerResponseWriter) StatusCode() int {
	return lrw.statusCode
}

func (lrw *loggerResponseWriter) Write(p []byte) (int, error) {
	if _, err := io.MultiWriter(lrw.responseBodyBuffer.Buffer, lrw.ResponseWriter).Write(p); err != nil {
		return 0, xerrors.Errorf("io.MultiWriter.Write: %w", err)
	}

	return len(p), nil
}

func AccessLogMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := rec.ContextLogger(r.Context())

			lrw, put := newLoggingResponseWriter(w, l)
			defer put(lrw.responseBodyBuffer)

			next.ServeHTTP(lrw, r)

			l.With(
				rec.Int("statusCode", lrw.StatusCode()),
				rec.Int64("contentLength", int64(lrw.responseBodyBuffer.Buffer.Len())),
			).F().Infof("access: %d %s (Content-Length:%d) <- %s %s (Content-Length:%d)", lrw.StatusCode(), http.StatusText(lrw.StatusCode()), lrw.responseBodyBuffer.Buffer.Len(), r.Method, r.URL.Path, r.ContentLength)
		})
	}
}
