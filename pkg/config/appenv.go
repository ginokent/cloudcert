package config

//go:generate go run golang.org/x/tools/cmd/stringer -trimprefix ENV -type=AppEnv appenv.go
type AppEnv int

const (
	ENVlocal AppEnv = iota
	ENVdevelop
)
