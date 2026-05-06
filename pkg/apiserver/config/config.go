package config

// Proxy configuration settings for the service proxy feature
type ProxyOptions struct {
	// Enabled indicates whether the service proxy feature is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`
	
	// RateLimitQPS indicates the maximum QPS for service proxy requests
	RateLimitQPS float32 `json:"rateLimitQPS" yaml:"rateLimitQPS"`
	
	// RateLimitBurst indicates the maximum burst of requests for service proxy
	RateLimitBurst int `json:"rateLimitBurst" yaml:"rateLimitBurst"`
}

// Add ProxyOptions to your existing server configuration struct
// For example:
/*
type Config struct {
    ...
    Proxy ProxyOptions `json:"proxy" yaml:"proxy"`
    ...
}
*/
