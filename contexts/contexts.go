package contexts

import (
	"bytes"
	"context"
	"time"

	"github.com/rec-logger/rec.go"
)

type contextKey string

const (
	ctxKeyLogger            contextKey = "logger"
	ctxKeyRequestBodyBuffer contextKey = "requestBodyBuffer"
	ctxKeyRequestTimeUTC    contextKey = "requestTimeUTC"
)

func WithLogger(ctx context.Context, l *rec.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, l)
}

func GetLogger(ctx context.Context) *rec.Logger {
	v := ctx.Value(ctxKeyLogger)
	if v == nil {
		return rec.L()
	}

	logger, ok := v.(*rec.Logger)
	if !ok {
		rec.F().Errorf("failed to type assertion: value=%#v, original type=%T, asserted type=%T", v, v, logger)

		return rec.L()
	}

	return logger
}

func WithRequestBodyBuffer(ctx context.Context, buf *bytes.Buffer) context.Context {
	return context.WithValue(ctx, ctxKeyRequestBodyBuffer, buf)
}

func GetRequestBodyBuffer(ctx context.Context) *bytes.Buffer {
	v := ctx.Value(ctxKeyRequestBodyBuffer)
	if v == nil {
		return bytes.NewBuffer(nil)
	}

	buf, ok := v.(*bytes.Buffer)
	if !ok {
		rec.F().Errorf("failed to type assertion: value=%#v, original type=%T, asserted type=%T", v, v, buf)

		return bytes.NewBuffer(nil)
	}

	return buf
}

func WithRequestTimeUTC(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, ctxKeyRequestTimeUTC, t)
}

func GetRequestTimeUTC(ctx context.Context) time.Time {
	v := ctx.Value(ctxKeyRequestTimeUTC)
	if v == nil {
		return time.Now().UTC()
	}

	t, ok := v.(time.Time)
	if !ok {
		rec.F().Errorf("failed to type assertion: value=%#v, original type=%T, asserted type=%T", v, v, t)

		return time.Now().UTC()
	}

	return t.UTC()
}
