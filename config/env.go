package config

import (
	"time"

	"github.com/newtstat/nits.go"
)

const (
	// APP_ENV environment variable name string.
	APP_ENV = "APP_ENV" // nolint: revive
	// ROLE environment variable name string.
	ROLE = "ROLE"
	// ADDR environment variable name string.
	ADDR = "ADDR"
	// PORT environment variable name string.
	PORT = "PORT"
	// GRPC_ENDPOINT environment variable name string.
	GRPC_ENDPOINT = "GRPC_ENDPOINT" // nolint: revive
	// SPAN_EXPORTER environment variable name string.
	SPAN_EXPORTER = "SPAN_EXPORTER" // nolint: revive
	// SHUTDOWN_TIMEOUT environment variable name string.
	SHUTDOWN_TIMEOUT = "SHUTDOWN_TIMEOUT" // nolint: revive
	// GOOGLE_CLOUD_PROJECT environment variable name string.
	GOOGLE_CLOUD_PROJECT = "GOOGLE_CLOUD_PROJECT" // nolint: revive
)

const (
	defaultAppEnv             = "default"
	defaultRole               = ""
	defaultAddr               = "0.0.0.0"
	defaultPort               = 8080
	defaultSpanExporter       = "local"
	defaultShutdownTimeout    = 10 * time.Second
	defaultGoogleCloudProject = ""
	defaultGRPCEndpoint       = "0.0.0.0:9090"
)

func loadEnv(c *config) *config {
	c.AppEnv = nits.Env.GetOrDefaultString(APP_ENV, defaultAppEnv)
	c.Role = nits.Env.GetOrDefaultString(ROLE, defaultRole)
	c.Addr = nits.Env.GetOrDefaultString(ADDR, defaultAddr)
	c.Port = nits.Env.GetOrDefaultInt(PORT, defaultPort)
	c.GRPCEndpoint = nits.Env.GetOrDefaultString(GRPC_ENDPOINT, defaultGRPCEndpoint)
	c.SpanExporter = nits.Env.GetOrDefaultString(SPAN_EXPORTER, defaultSpanExporter)
	c.ShutdownTimeout = nits.Env.GetOrDefaultSecond(SHUTDOWN_TIMEOUT, defaultShutdownTimeout)
	c.GoogleCloudProject = nits.Env.GetOrDefaultString(GOOGLE_CLOUD_PROJECT, defaultGoogleCloudProject)

	return c
}