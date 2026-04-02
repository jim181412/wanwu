package config

import (
	"os"
	"strings"
)

type Config struct {
	ContextPath  string
	AppServerURL string
	CASServerURL string
	AuthMode     string
	ListenAddr   string
}

func Load() Config {
	contextPath := normalizeContextPath(envOrDefault("CONTEXT_PATH", "/login"))
	casServerURL := normalizeCASServerURL(envOrDefault("CAS_SERVER_URL", "https://cas.ncu.edu.cn:8443/cas"))
	appServerURL := normalizeBaseURL(envOrDefault("APP_SERVER_URL", "http://auth.ncu.dilabs.cn"+contextPath))

	return Config{
		ContextPath:  contextPath,
		AppServerURL: appServerURL,
		CASServerURL: casServerURL,
		AuthMode:     "cas",
		ListenAddr:   envOrDefault("LISTEN_ADDR", ":9090"),
	}
}

func envOrDefault(key string, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return defaultValue
	}

	return value
}

func normalizeContextPath(path string) string {
	if path == "" {
		return "/go"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	path = strings.TrimRight(path, "/")
	if path == "" {
		return "/"
	}

	return path
}

func normalizeBaseURL(raw string) string {
	return strings.TrimRight(raw, "/")
}

func normalizeCASServerURL(raw string) string {
	baseURL := normalizeBaseURL(raw)

	switch {
	case strings.HasSuffix(baseURL, "/login"):
		return strings.TrimSuffix(baseURL, "/login")
	case strings.HasSuffix(baseURL, "/logout"):
		return strings.TrimSuffix(baseURL, "/logout")
	case strings.HasSuffix(baseURL, "/serviceValidate"):
		return strings.TrimSuffix(baseURL, "/serviceValidate")
	default:
		return baseURL
	}
}

func resolveAuthMode(raw string, casServerURL string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "cas":
		return "cas"
	case "mock":
		return "mock"
	}

	if casServerURL != "" {
		return "cas"
	}

	return "mock"
}
