# Практическое занятие №3: zap logging

Учебное HTTP-приложение на Go для практики структурированного логирования через `go.uber.org/zap`.

## Что реализовано

- `GET /health` — проверка работоспособности сервиса.
- `GET /students/{id}` — получение студента по идентификатору.
- Структурированные JSON-логи через zap.
- Middleware логирования HTTP-запросов.
- Логирование метода, пути, статуса, времени обработки, request_id и ошибок.
- Заголовок ответа `X-Request-ID`.
- Debug-логирование бизнес-операции поиска студента.

## Структура проекта

```text
pz3-logging/
  cmd/
    server/
      main.go
  internal/
    httpapi/
      handler.go
      middleware.go
      response_writer.go
    student/
      model.go
      repo.go
  pkg/
    logger/
      logger.go
  go.mod
```

## Запуск

```bash
go mod tidy
go run ./cmd/server
```

Сервер запускается на порту `8080`.

## Проверка

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:8080/students/1
curl -i http://localhost:8080/students/abc
curl -i http://localhost:8080/students/999
```
![Скриншот 1](./assets/Снимок%20экрана%C2%A0—%202026-05-03%20в%C2%A014.08.32.png)
----
![Скриншот 2](./assets/Снимок%20экрана%C2%A0—%202026-05-03%20в%C2%A014.09.19.png)
----
![Скриншот 3](./assets/Снимок%20экрана%C2%A0—%202026-05-03%20в%C2%A014.09.46.png)
----
![Скриншот 4](./assets/Снимок%20экрана%C2%A0—%202026-05-03%20в%C2%A014.10.03.png)
----
![Скриншот 5](./assets/Снимок%20экрана%C2%A0—%202026-05-03%20в%C2%A014.10.18.png)
----



## Уровень логирования

По умолчанию включён `debug`, чтобы в учебной практике были видны все уровни логов.

Можно изменить уровень через переменную окружения:

```bash
LOG_LEVEL=info go run ./cmd/server
```

Допустимые значения: `debug`, `info`, `warn`, `error`.
