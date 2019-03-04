package gamyeewai

import (
	"os"
)

const (
	EnvKeyAppEnv     = "APP_ENV"
	EnvKeyAppName    = "APP_NAME"
	EnvKeyAppRealEnv = "APP_REAL_ENV"

	EnvKeyGamyeewaiDingtalkToken = "GAMYEEWAI_DINGTALK_TOKEN"
	EnvKeyGamyeewaiKibanaDomain  = "GAMYEEWAI_KIBANA_DOMAIN"
	EnvKeyGamyeewaiKibanaIndex   = "GAMYEEWAI_KIBANA_INDEX"
)

func getAppEnv() string {
	if env, hasValue := os.LookupEnv(EnvKeyAppRealEnv); hasValue {
		return env
	}

	if env, hasValue := os.LookupEnv(EnvKeyAppEnv); hasValue {
		return env
	}

	return "local"
}

func getAppName() string {
	ret := getStringEnv(EnvKeyAppName, "gamyeewaiiii")

	return ret
}

func getStringEnv(key string, defaultValue ...string) string {
	if env, hasValue := os.LookupEnv(key); hasValue {
		return env
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}
