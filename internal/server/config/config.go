// Package config holds server service setup parameters.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// Config holds server service setup parameters.
type Config struct {
	// ConfigPath path to config file
	ConfigPath string `json:"config"`

	// address for the server to listen on
	ServerAddress string `json:"address"`

	// logger level
	LogLevel string `json:"log_level"`

	// Флаг -f, переменная окружения FILE_STORAGE_PATH — полное имя файла, куда
	// сохраняются текущие значения (по умолчанию /tmp/metrics-db.json, пустое
	// значение отключает функцию записи на диск).
	FileStoragePath string `json:"store_file"`

	// Флаг -i, переменная окружения STORE_INTERVAL — интервал времени в
	// секундах, по истечении которого текущие показания сервера сохраняются на
	// диск (по умолчанию 300 секунд, значение 0 делает запись синхронной)
	StoreInterval int `json:"store_interval"`

	// Флаг -r, переменная окружения RESTORE — булево значение (true/false),
	// определяющее, загружать или нет ранее сохранённые значения из указанного
	// файла при старте сервера (по умолчанию true).
	Restore bool `json:"restore"`

	// Строка с адресом подключения к БД должна получаться из переменной
	// окружения DATABASE_DSN или флага командной строки -d
	DatabaseDSN string `json:"database_dsn"`

	// Добавьте поддержку аргумента через флаг -k=<КЛЮЧ> и переменную
	// окружения KEY=<КЛЮЧ>.
	//  - При наличии ключа во время обработки запроса сервер должен проверять
	//    соответствие полученного и вычисленного хеша.
	//  - При несовпадении сервер должен отбрасывать полученные данные и
	//    возвращать http.StatusBadRequest.
	//  - При наличии ключа на этапе формирования ответа сервер должен вычислять
	//    хеш и передавать его в HTTP-заголовке ответа с именем HashSHA256.
	Key string `json:"key"`

	// CryptoKey is a public key to be used in messages encryption. Contains
	// path to a file with the key. Flag: -crypto-key, env: CRYPTO_KEY.
	//  > Шифруйте сообщения от агента к серверу с помощью ключей.
	CryptoKey string `json:"crypto_key"`
}

// New creates config with default values set.
func New() *Config {
	return &Config{
		ServerAddress:   "localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "/tmp/metrics-db.json",
		StoreInterval:   300,
		Restore:         true,
	}
}

// NewTesting returns config which is safe to use in tests.
func NewTesting() *Config {
	cfg := New()

	cfg.FileStoragePath = ""
	cfg.StoreInterval = -1
	cfg.Restore = false

	return cfg
}

// Sanitize checks for some input constraints and corrects if possible
// func (c *Config) Sanitize() (bool, error) {
// 	if c.StoreInterval < 0 {
// 		return false, errors.New("must be positive or zero")
// 	}

// 	return true, nil
// }

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
