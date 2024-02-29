// Package config holds agent service setup parameters.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// Config holds agent service setup parameters.
type Config struct {
	// ConfigPath path to config file
	ConfigPath string `json:"config"`

	// Server address for metrics to be sent to
	ServerURL string `json:"address"`

	// Interval in seconds
	PollInterval int `json:"poll_interval"`

	// Interval in seconds
	ReportInterval int `json:"report_interval"`

	// Shows if metrics update request should be sent in single batch
	Batch bool `json:"batch"`

	// Флаг -l, переменная окружения RATE_LIMIT.
	//   > количество одновременно исходящих запросов на сервер нужно ограничивать «сверху»
	RateLimit int `json:"rate_limit"`

	// Флаг -k=<КЛЮЧ> и переменная окружения KEY=<КЛЮЧ>.
	// При наличии ключа агент должен вычислять хеш и передавать
	// в HTTP-заголовке запроса с именем HashSHA256.
	Key string `json:"key"`

	// CryptoKey is a private key to be used in messages encryption. Contains
	// path to a file with the key. Flag: -crypto-key, env: CRYPTO_KEY.
	//  > Шифруйте сообщения от агента к серверу с помощью ключей.
	CryptoKey string `json:"crypto_key"`
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

// ParseFromFile parses config from file.
func ParseFromFile(cfg *Config, filepath string) error {
	if filepath = strings.TrimSpace(filepath); filepath == "" {
		return errors.New("filepath can't be empty")
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, cfg); err != nil {
		return err
	}

	return nil
}
