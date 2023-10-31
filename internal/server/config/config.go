package config

type Config struct {
	// address for the server to listen on
	ServerAddress string

	// logger level
	LogLevel string

	// Флаг -f, переменная окружения FILE_STORAGE_PATH — полное имя файла, куда
	// сохраняются текущие значения (по умолчанию /tmp/metrics-db.json, пустое
	// значение отключает функцию записи на диск).
	FileStoragePath string

	// Флаг -i, переменная окружения STORE_INTERVAL — интервал времени в
	// секундах, по истечении которого текущие показания сервера сохраняются на
	// диск (по умолчанию 300 секунд, значение 0 делает запись синхронной)
	StoreInterval int

	// Флаг -r, переменная окружения RESTORE — булево значение (true/false),
	// определяющее, загружать или нет ранее сохранённые значения из указанного
	// файла при старте сервера (по умолчанию true).
	Restore bool
}

// New creates config with default values set
func New() *Config {
	return &Config{
		ServerAddress:   "localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "/tmp/metrics-db.json",
		StoreInterval:   300,
		Restore:         true,
	}
}

// Sanitize checks for some input constraints and corrects if possible
// func (c *Config) Sanitize() (bool, error) {
// 	if c.StoreInterval < 0 {
// 		return false, errors.New("must be positive or zero")
// 	}

// 	return true, nil
// }
