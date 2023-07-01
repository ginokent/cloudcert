package config

import (
	"sync"

	configv1 "github.com/ginokent/cloudcert/proto/generated/go/config/v1"
	"github.com/kunitsuinc/util.go/env"
)

type config struct {
	AppEnv configv1.AppEnv
}

var (
	cfg   config
	cfgMu sync.Mutex
)

func Load() {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	cfg.AppEnv = configv1.AppEnv(env.IntOrDefault("APP_ENV", int(configv1.AppEnv_APP_ENV_LOCAL)))
}

func GetAppEnv() configv1.AppEnv {
	return cfg.AppEnv
}
