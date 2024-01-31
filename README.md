# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Testing

```Bash
go test ./... -coverprofile=profiles/cover.prof
go tool cover -func profiles/cover.prof
```

Or just print percentage only:
```Bash
go test ./... -coverprofile=profiles/cover.prof > /dev/null; go tool cover -func profiles/cover.prof | tail -n 1 | xargs
```

## Benchmarking

```Bash
go test -run=^$ -bench . ./internal/agent/ -memprofile=profiles/base.pprof -benchtime=400000x
```

```Bash
go test -run=^$ -bench . ./internal/agent/ -memprofile=profiles/result.pprof -benchtime=400000x
```

```Bash
go test -run=^$ -bench . ./internal/server/ -memprofile=profiles/mem-handlers.pprof -benchtime=125000x
```
