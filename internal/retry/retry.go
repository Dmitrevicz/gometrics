// Package retry implements retrier that allows to try again specific retriable
// errors.
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
	"errors"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"github.com/Dmitrevicz/gometrics/internal/model"
	"go.uber.org/zap"
)

// Retrier implements retry behaviour.
type Retrier struct {
	interval time.Duration
	retries  int
}

// NewRetrier creates new Retrier.
// Retrier will be limited by provided max retries and starting interval.
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

// progression decides what duration till next attempt should be waited
func (r *Retrier) progression(currentAttempt int) {
	if currentAttempt == 0 || currentAttempt > 5 {
		// interval duration will be increased at max 5 times
		// to prevent potentially endless wait
		return
	}

	r.interval = r.interval + time.Second
}

// Do does a retry of f(). Second bool param can be used to stop retries
// (e.g. when you need to attempt a retry only for some particular error
// but instantly return other errors without a retry).
func (r *Retrier) Do(action string, f func() error) (err error) {
	var retriable model.RetriableError
	for i := 0; i <= r.retries; i++ {
		if i > 0 {
			time.Sleep(r.interval)
			r.progression(i)
			logger.Log.Info("retrying...", zap.String("action", action), zap.Int("attempt", i))
		}

		if err = f(); err == nil || !errors.As(err, &retriable) {
			break // no need to try again
		}
	}

	return err
}
