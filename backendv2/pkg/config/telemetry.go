package config

type TelemetryConfig struct {
	Enabled     bool   `koanf:"enabled"`
	Exporter    string `koanf:"exporter"`    // "otlp" hopefully
	Endpoint    string `koanf:"endpoint"`    // OTLP endpoint URL
	ServiceName string `koanf:"serviceName"` // Override service name
	Insecure    bool   `koanf:"insecure"`    // Use HTTP instead of HTTPS

	Headers map[string]string `koanf:"headers"` // Additional headers for exporter

	// Distributed tracing context
	TraceParent string `koanf:"traceParent"`
	TraceState  string `koanf:"traceState"`
}
