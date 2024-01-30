// Package config holds server service setup parameters.
package config

// Config holds server service setup parameters.
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

	// Строка с адресом подключения к БД должна получаться из переменной
	// окружения DATABASE_DSN или флага командной строки -d
	DatabaseDSN string

	// Добавьте поддержку аргумента через флаг -k=<КЛЮЧ> и переменную
	// окружения KEY=<КЛЮЧ>.
	//  - При наличии ключа во время обработки запроса сервер должен проверять
	//    соответствие полученного и вычисленного хеша.
	//  - При несовпадении сервер должен отбрасывать полученные данные и
	//    возвращать http.StatusBadRequest.
	//  - При наличии ключа на этапе формирования ответа сервер должен вычислять
	//    хеш и передавать его в HTTP-заголовке ответа с именем HashSHA256.
	Key string
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
