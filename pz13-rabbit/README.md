# Практическое занятие №13 — RabbitMQ + Go

Проект демонстрирует асинхронное взаимодействие через RabbitMQ:

1. клиент вызывает `POST /v1/tasks`;
2. сервис `tasks` создаёт задачу;
3. сервис `tasks` публикует событие `task.created` в очередь `task_events`;
4. отдельный `worker` получает сообщение;
5. worker подтверждает обработку через `ack` и пишет событие в лог.

## Структура проекта

```text
pz13-rabbit/
├── deploy/
│   └── rabbit/
│       └── docker-compose.yml
├── internal/
│   ├── amqpclient/
│   │   └── connect.go
│   └── events/
│       └── task_event.go
├── services/
│   ├── tasks/
│   │   ├── cmd/tasks/main.go
│   │   └── internal/
│   │       ├── config/config.go
│   │       ├── httpapi/server.go
│   │       ├── publisher/publisher.go
│   │       └── service/task_service.go
│   └── worker/
│       ├── cmd/worker/main.go
│       └── internal/
│           ├── config/config.go
│           └── consumer/consumer.go
├── .gitignore
├── go.mod
└── README.md
```

## Требования

Для запуска на macOS нужны:

- Docker Desktop;
- Go 1.22 или новее;
- терминал macOS.

Проверка:

```bash
docker --version
docker compose version
go version
```

## Команды от А до Я

### 1. Перейти в папку проекта

```bash
cd pz13-rabbit
```

### 2. Загрузить Go-зависимости

```bash
go mod tidy
```

### 3. Проверить компиляцию проекта

```bash
go test ./...
```

### 4. Запустить RabbitMQ

```bash
cd deploy/rabbit
docker compose up -d
docker compose ps
cd ../..
```

RabbitMQ Management UI будет доступен по адресу:

```text
http://localhost:15672
```

Логин и пароль:

```text
guest / guest
```

### 5. Запустить worker в отдельном терминале

Откройте новый терминал, перейдите в папку проекта и выполните:

```bash
cd pz13-rabbit
export RABBIT_URL="amqp://guest:guest@localhost:5672/"
export QUEUE_NAME="task_events"
export WORKER_PREFETCH="1"
go run ./services/worker/cmd/worker
```

Ожидаемый лог:

```text
worker: ... worker started: queue=task_events prefetch=1
```

### 6. Запустить сервис tasks во втором отдельном терминале

Откройте ещё один терминал, перейдите в папку проекта и выполните:

```bash
cd pz13-rabbit
export RABBIT_URL="amqp://guest:guest@localhost:5672/"
export QUEUE_NAME="task_events"
export TASKS_ADDR=":8082"
go run ./services/tasks/cmd/tasks
```

Ожидаемый лог:

```text
tasks: ... rabbit connected: queue=task_events
tasks: ... HTTP server started: addr=:8082
```

### 7. Создать задачу через HTTP API

Откройте третий терминал и выполните:

```bash
curl -i -X POST http://localhost:8082/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -H "X-Request-ID: pz13-001" \
  -d '{"title":"Rabbit","description":"publish event"}'
```

Ожидаемый HTTP-ответ:

```text
HTTP/1.1 201 Created
Content-Type: application/json
```

Пример тела ответа:

```json
{
  "event_publish_status": "published",
  "request_id": "pz13-001",
  "task": {
    "id": "t_001",
    "title": "Rabbit",
    "description": "publish event",
    "created_at": "2026-03-26T10:00:00Z"
  }
}
```

### 8. Проверить лог worker

В терминале worker должна появиться строка вида:

```text
worker: ... received event=task.created task_id=t_001 ts=... request_id=pz13-001 producer=tasks version=1
```

Это подтверждает, что сообщение прошло путь:

```text
HTTP request -> tasks service -> RabbitMQ queue -> worker -> ack
```

### 9. Проверить очередь в RabbitMQ Management UI

1. Откройте `http://localhost:15672`.
2. Войдите под `guest / guest`.
3. Перейдите во вкладку `Queues and Streams`.
4. Найдите очередь `task_events`.

### 10. Дополнительная проверка списка задач

```bash
curl -i http://localhost:8082/v1/tasks
```

### 11. Остановить сервисы

В терминалах `worker` и `tasks` нажмите:

```text
Control + C
```

### 12. Остановить RabbitMQ

```bash
cd deploy/rabbit
docker compose down
cd ../..
```

## Переменные окружения

| Переменная | Значение по умолчанию | Назначение |
|---|---:|---|
| `RABBIT_URL` | `amqp://guest:guest@localhost:5672/` | URL подключения к RabbitMQ |
| `QUEUE_NAME` | `task_events` | Имя очереди |
| `WORKER_PREFETCH` | `1` | Количество неподтверждённых сообщений у worker |
| `TASKS_ADDR` | `:8082` | Адрес HTTP-сервера tasks |

## Формат события

```json
{
  "event": "task.created",
  "task_id": "t_001",
  "ts": "2026-03-26T10:00:00Z",
  "request_id": "pz13-001",
  "producer": "tasks",
  "version": "1"
}
```

## Что реализовано

- RabbitMQ через `docker compose`;
- очередь `task_events` с `durable = true`;
- HTTP-сервис `tasks`;
- публикация события `task.created` после создания задачи;
- отдельный worker-процесс;
- ручной `ack` после успешной обработки;
- `prefetch = 1`;

## Скриншоты
![Скриншот 1](./assets/scr_1.png)
----
![Скриншот 2](./assets/scr_2.png)
----
![Скриншот 3](./assets/scr_3.png)
----
![Скриншот 3](./assets/scr_4.png)
----
![Скриншот 3](./assets/scr_5.png)
----
![Скриншот 3](./assets/scr_6.png)

# Контрольные вопросы — Практическое занятие №13

## 1. Зачем нужен брокер сообщений, если между сервисами можно использовать HTTP?

HTTP удобен для синхронных запросов, когда вызывающий сервис сразу ждёт ответ. Брокер сообщений нужен для асинхронных задач: сервис публикует событие в очередь и быстро отвечает клиенту, а отдельный обработчик выполняет работу позже. Это снижает связанность сервисов и помогает переживать временную недоступность обработчика.

## 2. Что такое producer и consumer?

Producer — компонент, который публикует сообщение в очередь. В этой работе producer — сервис `tasks`, который отправляет событие `task.created`.

Consumer — компонент, который читает сообщения из очереди и обрабатывает их. В этой работе consumer — отдельный `worker`, который читает очередь `task_events`.

## 3. Почему сообщение нужно публиковать только после успешного создания задачи?

Если отправить событие до создания задачи, worker может получить сообщение о задаче, которой фактически нет. Поэтому сначала нужно создать задачу, получить её `task_id`, а уже затем публиковать событие `task.created`.

## 4. Что такое ack и зачем он нужен?

`ack` — подтверждение успешной обработки сообщения consumer-ом. После `Ack(false)` RabbitMQ считает сообщение обработанным и удаляет его из очереди.

## 5. Почему возможна повторная доставка сообщения?

Повторная доставка возможна, если consumer получил сообщение, но не успел отправить `ack`: например, процесс упал или соединение с RabbitMQ оборвалось. Тогда брокер может снова выдать это сообщение другому consumer-у или тому же consumer-у после восстановления.

## 6. Что делает параметр prefetch?

`prefetch` ограничивает количество сообщений, которые RabbitMQ может передать consumer-у без подтверждения. В проекте используется `prefetch = 1`, чтобы worker обрабатывал по одному неподтверждённому сообщению и не забирал из очереди слишком много задач заранее.

## 7. Чем durable queue отличается от non-durable queue?

`durable` очередь сохраняет своё объявление при перезапуске RabbitMQ. `non-durable` очередь после перезапуска брокера исчезает. В проекте очередь `task_events` объявлена как durable.

## 8. Почему JSON удобен как учебный формат сообщения?

JSON легко читать в логах, удобно сериализовать и десериализовать стандартным пакетом Go `encoding/json`, а также он хорошо подходит для простого обмена структурированными данными между сервисами.

## 9. Что означает режим best effort при публикации события?

`best effort` означает, что задача считается созданной даже при ошибке публикации события. Ошибка отправки в RabbitMQ логируется, но клиент всё равно получает успешный HTTP-ответ о создании задачи.

## 10. Почему worker лучше выносить в отдельный процесс, а не смешивать с HTTP-обработчиками?

Отдельный worker проще масштабировать, перезапускать и контролировать независимо от HTTP API. Если обработка сообщений станет долгой или нестабильной, она не будет напрямую блокировать HTTP-обработчики сервиса `tasks`.
