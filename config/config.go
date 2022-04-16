package config

import (
	"runtime/debug"
	"time"
)

type config struct {
	AppEnv             string
	Role               string
	Addr               string
	Port               int
	GRPCEndpoint       string
	SpanExporter       string
	ShutdownTimeout    time.Duration
	GoogleCloudProject string
	GoVersion          string
	Version            string
	Revision           string
	Branch             string
	Timestamp          string
}

// nolint: gochecknoglobals
var c = loadEnv(loadDefault())

func AppEnv() string                 { return c.AppEnv }
func Role() string                   { return c.Role }
func Addr() string                   { return c.Addr }
func Port() int                      { return c.Port }
func GRPCEndpoint() string           { return c.GRPCEndpoint }
func SpanExporter() string           { return c.SpanExporter }
func ShutdownTimeout() time.Duration { return c.ShutdownTimeout }
func GoogleCloudProject() string     { return c.GoogleCloudProject }
func GoVersion() string              { return c.GoVersion }
func Version() string                { return c.Version }
func Revision() string               { return c.Revision }
func Branch() string                 { return c.Branch }
func Timestamp() string              { return c.Timestamp }

func loadDefault() *config {
	cfg := &config{
		Version:   version,
		Revision:  revision,
		Branch:    branch,
		Timestamp: timestamp,
	}

	info, ok := debug.ReadBuildInfo()
	if ok {
		cfg.GoVersion = info.GoVersion
	}

	return cfg
}
