package config

import "os"

type Config struct {
	Addr            string
	Domain          string
	DataDir         string
	CredentialsFile string
	SessionSecret   string
}

func Load() *Config {
	return &Config{
		Addr:            envOr("XZAR_ADDR", ":8080"),
		Domain:          envOr("XZAR_DOMAIN", "xz.ar"),
		DataDir:         envOr("XZAR_DATA_DIR", "/data"),
		CredentialsFile: envOr("XZAR_CREDENTIALS_FILE", "/etc/xzar/credentials.json"),
		SessionSecret:   envOr("XZAR_SESSION_SECRET", ""),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
