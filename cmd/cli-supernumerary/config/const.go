package config

const (
	EnvBundleID  = "CFG_BUNDLE_ID"
	EnvNatsUrl   = "CFG_NATS_URL"
	EnvRedisDSN  = "CFG_CACHE_REDIS_PERM_DSN"
	EnvRedisPass = "CFG_CACHE_REDIS_PERM_PASS"
)

const (
	SubjectCommander     = "SUBJECT_COMMANDER"
	SubjectStart         = "SUBJECT_START"
	SubjectStop          = "SUBJECT_STOP"
	SubjectTicker        = "SUBJECT_TICKER"
	SubjectCreateAuthKey = "SUBJECT_CREATE_AUTH_KEY"
	SubjectRegister      = "SUBJECT_REGISTER"
	SubjectLogin         = "SUBJECT_LOGIN"
	SubjectPhoneRange    = "SUBJECT_PHONE_RANGE"
	SubjectHealthCheck   = "SUBJECT_HEALTH_CHECK"
	SubjectResetAuth     = "SUBJECT_RESET_AUTH"
)
