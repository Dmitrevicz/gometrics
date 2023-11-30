package config

type Config struct {
	// Server address for metrics to be sent to
	ServerURL string

	// Interval in seconds
	PollInterval int

	// Interval in seconds
	ReportInterval int

	// Shows if metrics update request should be sent in single batch
	Batch bool

	// Флаг -k=<КЛЮЧ> и переменная окружения KEY=<КЛЮЧ>.
	// При наличии ключа агент должен вычислять хеш и передавать
	// в HTTP-заголовке запроса с именем HashSHA256.
	Key string
}

// New creates config with default values set
func New() *Config {
	return &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   2,
		ReportInterval: 10,
		Batch:          true,
	}
}
