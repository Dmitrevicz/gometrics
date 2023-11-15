// Package retry.
//
// Задание инкремента #13
//
// Измените весь свой код в соответствии со знаниями, полученными в этой теме.
// Добавьте обработку retriable-ошибок.
//
// Retriable-ошибки — это ошибки, которые могут быть исправлены повторной
// попыткой выполнения операции. Это бывает полезно для программ, которые
// работают с сетью или файловой системой, где возможны временные проблемы связи
// или доступа к данным. Ошибки могут быть вызваны различными причинами, такими
// как перегрузка сервера, недоступность сети или ошибки в коде программы.
//
// Примеры retriable-ошибок:
//   - Ошибка связи с сервером при отправке запроса.
//   - Ошибка чтения данных из сети или БД из-за проблем соединения.
//   - Ошибка доступа к файлу, который был заблокирован другим процессом.
//
// Сценарии возможных ошибок:
//   - Агент не сумел с первой попытки выгрузить данные на сервер из-за временной
//     невозможности установить соединение с сервером.
//   - При обращении к PostgreSQL cервер получил ошибку транспорта (из категории
//     Class 08 — Connection Exception).
//
// Стратегия реализации:
//   - Количество повторов должно быть ограничено тремя дополнительными попытками.
//   - Интервалы между повторами должны увеличиваться: 1s, 3s, 5s.
//   - Чтобы определить тип ошибки PostgreSQL, с которой завершился запрос, можно
//     воспользоваться библиотекой github.com/jackc/pgerrcode, в частности
//     pgerrcode.UniqueViolation.
package retry

import (
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"go.uber.org/zap"
)

type Retrier struct {
	interval time.Duration
	retries  int
	i        int
}

func NewRetrier(interval time.Duration, retries int) *Retrier {
	if interval <= 0 {
		interval = time.Second
	}

	if retries < 0 {
		retries = 0
	}

	r := Retrier{
		interval: interval,
		retries:  retries,
	}

	return &r
}

func (r *Retrier) progression() {
	if r.i == 0 || r.i > 5 {
		return
	}

	r.interval = r.interval + time.Second
}

// Do does a retry of f(). Second bool param can be used to stop retries
// (e.g. when you need to attempt a retry only for some particular error
// but instantly return other errors without a retry).
func (r *Retrier) Do(action string, f func() (error, bool)) (err error) {
	var try bool
	for r.i = 0; r.i <= r.retries; r.i++ {
		if r.i > 0 {
			time.Sleep(r.interval)
			r.progression()
			logger.Log.Info("retrying...", zap.String("action", action), zap.Int("attempt", r.i))
		}

		if err, try = f(); err == nil || !try {
			break
		}
	}

	return err
}
