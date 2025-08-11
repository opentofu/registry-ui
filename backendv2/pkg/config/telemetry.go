package config

type TelemetryConfig struct {
	Enabled     bool   `koanf:"enabled"`
	Exporter    string `koanf:"exporter"`    // "otlp" or "jaeger" or "stdout"
	Endpoint    string `koanf:"endpoint"`    // OTLP endpoint URL
	ServiceName string `koanf:"serviceName"` // Override service name
	Insecure    bool   `koanf:"insecure"`    // Use HTTP instead of HTTPS

	// Distributed tracing context
	TraceParent string `koanf:"traceParent"`
	TraceState  string `koanf:"traceState"`
}
