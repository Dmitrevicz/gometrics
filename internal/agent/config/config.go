// Package config holds agent service setup parameters.
package config

// Config holds agent service setup parameters.
type Config struct {
	// Server address for metrics to be sent to
	ServerURL string

	// Interval in seconds
	PollInterval int

	// Interval in seconds
	ReportInterval int

	// Shows if metrics update request should be sent in single batch
	Batch bool

	// Флаг -l, переменная окружения RATE_LIMIT.
	//   > количество одновременно исходящих запросов на сервер нужно ограничивать «сверху»
	RateLimit int

	// Флаг -k=<КЛЮЧ> и переменная окружения KEY=<КЛЮЧ>.
	// При наличии ключа агент должен вычислять хеш и передавать
	// в HTTP-заголовке запроса с именем HashSHA256.
	Key string

	// CryptoKey is a private key to be used in messages encryption. Contains
	// path to a file with the key. Flag: -crypto-key, env: CRYPTO_KEY.
	//  > Шифруйте сообщения от агента к серверу с помощью ключей.
	CryptoKey string
}

// New creates config with default values set.
func New() *Config {
	return &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   2,
		ReportInterval: 10,
		Batch:          true,
		RateLimit:      1,
	}
}
