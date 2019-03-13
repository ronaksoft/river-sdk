package config

import "time"

type StartCfg struct {
	ServerURL     string        `json:"server-url"`
	FileServerURL string        `json:"file-server-url"`
	Timeout       time.Duration `json:"timeout"`
}
