# Практическое занятие №15 — деплой Go-приложения на VPS через systemd

Проект `tasks` — учебный backend-сервис на Go для практики деплоя на Linux VPS. Сервис запускается как systemd-служба, читает конфигурацию из env-файла и имеет endpoint проверки состояния `GET /health`.

## Что реализовано

- Go HTTP-сервис без внешних зависимостей.
- Endpoint `GET /health` для проверки работоспособности.
- Простые in-memory endpoints для задач:
  - `GET /tasks`
  - `POST /tasks`
  - `GET /tasks/{id}`
  - `PATCH /tasks/{id}`
  - `DELETE /tasks/{id}`
- Конфигурация через переменные окружения.
- Готовый пример `configs/tasks.env.example`.
- Готовый systemd unit-файл `deploy/tasks.service`.
- Команды для сборки на macOS, переноса бинарника на VPS, запуска, проверки логов, обновления и отката.

## Структура проекта

```text
tasks-systemd-practice/
├── cmd/tasks/main.go
├── configs/tasks.env.example
├── deploy/tasks.service
├── internal/config/config.go
├── internal/config/config_test.go
├── internal/server/server.go
├── internal/server/server_test.go
├── .gitignore
├── go.mod
└── README.md
```

## Переменные окружения

| Переменная | Пример | Назначение |
|---|---:|---|
| `TASKS_HOST` | `0.0.0.0` | IP-адрес, на котором слушает приложение |
| `TASKS_PORT` | `8082` | Порт приложения |
| `AUTH_BASE_URL` | `http://127.0.0.1:8081` | URL внешнего auth-сервиса, если он нужен |
| `REDIS_ADDR` | `127.0.0.1:6379` | Адрес Redis, если он нужен |
| `LOG_LEVEL` | `info` | Уровень логирования; для тестов можно использовать `silent` |
| `READ_TIMEOUT_SECONDS` | `5` | Таймаут чтения HTTP-запроса |
| `WRITE_TIMEOUT_SECONDS` | `10` | Таймаут записи HTTP-ответа |
| `IDLE_TIMEOUT_SECONDS` | `60` | Таймаут idle-соединения |

## Команды от А до Я


### 1. Подготовить проект на macOS

Проверить Go:

```bash
go version
```

Проверить тесты:

```bash
go test ./...
```

Запустить локально:

```bash
TASKS_HOST=127.0.0.1 TASKS_PORT=8082 go run ./cmd/tasks
```

В другом терминале проверить `/health`:

```bash
curl -i http://127.0.0.1:8082/health
```

Ожидаемый ответ:

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8

{"service":"tasks","status":"ok"}
```

Пример проверки создания задачи:

```bash
curl -i -X POST http://127.0.0.1:8082/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"deploy app to VPS"}'
```

Получить список задач:

```bash
curl -i http://127.0.0.1:8082/tasks
```

Остановить локальный сервис: нажмите `Ctrl+C` в терминале, где выполняется `go run`.

### 2. Собрать Linux-бинарник на macOS

Для VPS с обычной x86_64/amd64 архитектурой:

```bash
mkdir -p bin
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/tasks ./cmd/tasks
```

Проверить, что файл создан:

```bash
ls -lh bin/tasks
file bin/tasks
```

Если VPS на ARM64, например Ampere/Graviton, используйте другую сборку:

```bash
mkdir -p bin
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o bin/tasks ./cmd/tasks
```

### 3. Подключиться к VPS

```bash
ssh user@<VPS_IP>
```

### 4. Обновить пакеты на VPS

```bash
sudo apt update && sudo apt upgrade -y
```

### 5. Создать системного пользователя для сервиса

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin tasksuser
```

### 6. Создать директорию приложения на VPS

```bash
sudo mkdir -p /opt/tasks
sudo chown -R tasksuser:tasksuser /opt/tasks
```

### 7. Создать директорию конфигурации на VPS

```bash
sudo mkdir -p /etc/tasks
```

Создать env-файл:

```bash
sudo nano /etc/tasks/tasks.env
```

Вставить содержимое:

```env
TASKS_HOST=0.0.0.0
TASKS_PORT=8082
AUTH_BASE_URL=http://127.0.0.1:8081
REDIS_ADDR=127.0.0.1:6379
LOG_LEVEL=info
READ_TIMEOUT_SECONDS=5
WRITE_TIMEOUT_SECONDS=10
IDLE_TIMEOUT_SECONDS=60
```

Сохранить файл в `nano`: `Ctrl+O`, `Enter`, `Ctrl+X`.

Назначить безопасные права:

```bash
sudo chown root:root /etc/tasks/tasks.env
sudo chmod 600 /etc/tasks/tasks.env
```

### 8. Скопировать бинарник с macOS на VPS

На macOS, из папки проекта:

```bash
scp bin/tasks user@<VPS_IP>:/tmp/tasks
```

### 9. Переместить бинарник в рабочую директорию на VPS

На VPS:

```bash
sudo mv /tmp/tasks /opt/tasks/tasks
sudo chown tasksuser:tasksuser /opt/tasks/tasks
sudo chmod 755 /opt/tasks/tasks
```

### 10. Создать systemd unit-файл

На VPS:

```bash
sudo nano /etc/systemd/system/tasks.service
```

Вставить:

```ini
[Unit]
Description=Tasks Service
After=network.target

[Service]
Type=simple
User=tasksuser
WorkingDirectory=/opt/tasks
EnvironmentFile=/etc/tasks/tasks.env
ExecStart=/opt/tasks/tasks
Restart=always
RestartSec=2
NoNewPrivileges=true
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

Сохранить файл: `Ctrl+O`, `Enter`, `Ctrl+X`.

### 11. Перечитать конфигурацию systemd

```bash
sudo systemctl daemon-reload
```

### 12. Запустить сервис

```bash
sudo systemctl start tasks
```

### 13. Включить автозапуск

```bash
sudo systemctl enable tasks
```

### 14. Проверить статус сервиса

```bash
sudo systemctl status tasks
```

Служба должна быть в состоянии `active (running)`.

### 15. Посмотреть логи

Последние 100 строк:

```bash
sudo journalctl -u tasks --no-pager -n 100
```

Логи в реальном времени:

```bash
sudo journalctl -u tasks -f
```

### 16. Проверить `/health` на VPS

На VPS:

```bash
curl -i http://127.0.0.1:8082/health
```

Если порт открыт наружу, можно проверить с macOS:

```bash
curl -i http://<VPS_IP>:8082/health
```

### 17. Базовые команды управления сервисом

Запуск:

```bash
sudo systemctl start tasks
```

Остановка:

```bash
sudo systemctl stop tasks
```

Перезапуск:

```bash
sudo systemctl restart tasks
```

Статус:

```bash
sudo systemctl status tasks
```

Отключить автозапуск:

```bash
sudo systemctl disable tasks
```

Включить автозапуск:

```bash
sudo systemctl enable tasks
```


## Контрольные вопросы и ответы

### 1. Что такое VPS и зачем он нужен backend-разработчику?

VPS — это виртуальный сервер, на котором можно запускать backend-приложения, настраивать окружение, управлять сервисами и делать приложение доступным из сети. Backend-разработчику VPS нужен для публикации и эксплуатации сервисов вне локального компьютера.

### 2. Почему запуск приложения на VPS отличается от локального запуска на компьютере разработчика?

Локально приложение обычно запускается вручную из терминала или IDE. На VPS приложение должно работать постоянно, переживать закрытие SSH-сессии, автоматически запускаться после перезагрузки и иметь централизованные логи. Поэтому используется systemd или другой менеджер процессов.

### 3. Для чего используется systemd?

systemd используется для управления Linux-службами: запуска, остановки, перезапуска, автозапуска при старте сервера, контроля состояния и просмотра логов через journalctl.

### 4. Почему не рекомендуется запускать серверное приложение от root?

Запуск от root повышает риск компрометации всей системы. Если в приложении будет уязвимость, злоумышленник получит максимальные права. Безопаснее запускать сервис от отдельного системного пользователя с минимальными правами.

### 5. Зачем выносить конфигурацию в отдельный env-файл?

Env-файл позволяет менять настройки без перекомпиляции приложения, не хранить секреты в коде, использовать один и тот же бинарник в разных окружениях и упростить эксплуатацию.

### 6. Что делает параметр `Restart=always`?

`Restart=always` говорит systemd перезапускать сервис после завершения процесса. Это помогает восстановить работу приложения после сбоя.

### 7. Для чего нужен `EnvironmentFile` в unit-файле?

`EnvironmentFile` подключает файл с переменными окружения. Эти переменные становятся доступны процессу приложения при запуске через systemd.

### 8. Как проверить состояние службы через systemctl?

```bash
sudo systemctl status tasks
```

Эта команда показывает, запущена ли служба, её PID, последние сообщения и возможные ошибки запуска.

### 9. Как посмотреть логи сервиса через journalctl?

```bash
sudo journalctl -u tasks --no-pager -n 100
```

Для просмотра логов в реальном времени:

```bash
sudo journalctl -u tasks -f
```

### 10. Что нужно сделать перед обновлением unit-файла systemd?

После изменения unit-файла нужно выполнить:

```bash
sudo systemctl daemon-reload
```

Затем можно перезапустить сервис:

```bash
sudo systemctl restart tasks
```

### 11. Почему полезно иметь процедуру отката версии?

Откат нужен, чтобы быстро вернуть предыдущую рабочую версию, если новая версия не запускается, содержит ошибку или нарушает работу сервиса.

### 12. Зачем в реальных системах часто используют NGINX перед приложением?

NGINX обычно принимает внешний HTTP/HTTPS-трафик, завершает TLS, проксирует запросы к backend-приложению, скрывает внутренний порт приложения, может выполнять сжатие, rate limiting и отдачу статических файлов.
