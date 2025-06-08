package logger

import "go.uber.org/zap"

func New(logLevel string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		return nil, err
	}
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = lvl
	logCfg.DisableCaller = true
	logCfg.DisableStacktrace = true

	return logCfg.Build()
}
