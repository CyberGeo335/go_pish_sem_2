# Практическое занятие №14 — очередь задач producer–consumer на Go и RabbitMQ

Проект реализует учебную очередь задач по модели **producer–consumer**:

- HTTP-сервис `tasks` принимает запрос на постановку задачи.
- `tasks` публикует JSON-сообщение в очередь RabbitMQ `task_jobs`.
- `worker` читает задачи из `task_jobs`, имитирует тяжёлую обработку и подтверждает сообщение через `ack`.
- При ошибке `worker` выполняет ограниченные retries.
- После исчерпания попыток сообщение отправляется в `task_jobs_dlq`.
- Для учебной идемпотентности `worker` хранит уже обработанные `message_id` в памяти.

---

## 1. Структура проекта

```text
pz14-task-queue/
├── deploy/
│   └── rabbit/
│       └── docker-compose.yml
├── internal/
│   ├── jobs/
│   │   └── task_job.go
│   └── rabbitmq/
│       └── rabbitmq.go
├── services/
│   ├── tasks/
│   │   ├── cmd/
│   │   │   └── tasks/
│   │   │       └── main.go
│   │   └── internal/
│   │       └── http/
│   │           └── handler.go
│   └── worker/
│       ├── cmd/
│       │   └── worker/
│       │       └── main.go
│       └── internal/
│           ├── consumer/
│           │   └── consumer.go
│           └── store/
│               └── store.go
├── .gitignore
├── go.mod
└── README.md
```

---

## 2. Что реализовано

### Очереди

| Очередь | Назначение |
|---|---|
| `task_jobs` | Основная очередь задач |
| `task_jobs_dlq` | Dead Letter Queue для проблемных сообщений |

### Формат сообщения

```json
{
  "job": "process_task",
  "task_id": "t_001",
  "attempt": 1,
  "message_id": "uuid-here"
}
```

Поля:

- `job` — тип задачи.
- `task_id` — идентификатор бизнес-объекта.
- `attempt` — номер текущей попытки обработки.
- `message_id` — уникальный идентификатор сообщения для идемпотентной проверки.

### HTTP endpoint

```text
POST /v1/jobs/process-task
```

Тело запроса:

```json
{
  "task_id": "t_001"
}
```

Успешный ответ:

```json
{
  "status": "accepted",
  "task_id": "t_001",
  "message_id": "generated-uuid",
  "queue": "task_jobs"
}
```

### Правило имитации ошибки

В `worker` используется простое учебное правило:

- `task_id = "t_fail"` — задача всегда завершается ошибкой;
- любой другой `task_id` — задача завершается успешно.

Максимальное число попыток: `3`.

---

## 3. Требования для macOS

Нужно установить:

- Go;
- Docker Desktop;
- curl.

Проверка Go:

```bash
go version
```

## 4. Запуск проекта от А до Я на macOS

### Шаг 1. Скачать Go-зависимости

```bash
go mod download
```

Дополнительно можно привести зависимости в порядок:

```bash
go mod tidy
```

### Шаг 2. Проверить компиляцию проекта

```bash
go test ./...
```

### Шаг 3. Поднять RabbitMQ

```bash
cd ~/projects/pz14-task-queue/deploy/rabbit
```

```bash
docker compose up -d
```

```bash
docker compose ps
```

Проверить логи контейнера:

```bash
docker logs -f pz14-rabbitmq
```

Остановить просмотр логов можно сочетанием клавиш `Control + C`.

RabbitMQ Management UI:

```text
http://localhost:15672
```

Логин:

```text
guest
```

Пароль:

```text
guest
```

### Шаг 5. Запустить worker

Откройте новый терминал.

```bash
cd ~/projects/pz14-task-queue
```

```bash
export RABBIT_URL="amqp://guest:guest@localhost:5672/"
```

```bash
go run ./services/worker/cmd/worker
```

Ожидаемый лог:

```text
worker is consuming queue=task_jobs max_attempts=3
```

### Шаг 6. Запустить HTTP-сервис tasks

Откройте ещё один новый терминал.

```bash
cd ~/projects/pz14-task-queue
```

```bash
export RABBIT_URL="amqp://guest:guest@localhost:5672/"
```

```bash
export AUTH_TOKEN="demo-token"
```

```bash
export HTTP_ADDR=":8082"
```

```bash
go run ./services/tasks/cmd/tasks
```

Ожидаемый лог:

```text
tasks service listening on :8082
POST /v1/jobs/process-task publishes messages to task_jobs
```

### Шаг 7. Проверить health endpoint

Откройте ещё один терминал.

```bash
curl -i http://localhost:8082/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

---

## 5. Проверка успешной обработки

Отправить обычную задачу:

```bash
curl -i -X POST http://localhost:8082/v1/jobs/process-task \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -d '{"task_id":"t_001"}'
```

Ожидаемый HTTP-ответ:

```text
HTTP/1.1 202 Accepted
```

Пример JSON-ответа:

```json
{
  "status": "accepted",
  "task_id": "t_001",
  "message_id": "generated-uuid",
  "queue": "task_jobs"
}
```

Ожидаемое поведение worker:

```text
received job=process_task task_id=t_001 ... attempt=1
simulate heavy work task_id=t_001
processed successfully task_id=t_001 ...; ack
```

---

## 6. Проверка retries и DLQ

Отправить задачу, которая всегда падает:

```bash
curl -i -X POST http://localhost:8082/v1/jobs/process-task \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -d '{"task_id":"t_fail"}'
```

Ожидаемое поведение:

1. Первая попытка завершается ошибкой.
2. Сообщение публикуется обратно в `task_jobs` с `attempt = 2`.
3. Вторая попытка завершается ошибкой.
4. Сообщение публикуется обратно в `task_jobs` с `attempt = 3`.
5. Третья попытка завершается ошибкой.
6. Сообщение публикуется в `task_jobs_dlq`.
7. Исходное сообщение подтверждается через `ack`, чтобы не было бесконечного цикла.

Пример логов worker:

```text
received job=process_task task_id=t_fail ... attempt=1
processing failed task_id=t_fail ... attempt=1 error=simulated processing error
retry task_id=t_fail ... next_attempt=2
received job=process_task task_id=t_fail ... attempt=2
processing failed task_id=t_fail ... attempt=2 error=simulated processing error
retry task_id=t_fail ... next_attempt=3
received job=process_task task_id=t_fail ... attempt=3
processing failed task_id=t_fail ... attempt=3 error=simulated processing error
max attempts reached task_id=t_fail ... attempt=3; publish to task_jobs_dlq
```

---

## 7. Проверка DLQ через RabbitMQ Management UI

Откройте:

```text
http://localhost:15672
```

Дальше:

1. Войти под `guest / guest`.
2. Открыть вкладку **Queues and Streams**.
3. Найти очередь `task_jobs`.
4. Найти очередь `task_jobs_dlq`.
5. Убедиться, что после запроса с `task_id = "t_fail"` в `task_jobs_dlq` появилось сообщение.

Проверить DLQ через HTTP API RabbitMQ:

```bash
curl -s -u guest:guest http://localhost:15672/api/queues/%2F/task_jobs_dlq | python3 -m json.tool
```

Посмотреть сообщения из DLQ без удаления из очереди:

```bash
curl -s -u guest:guest \
  -H "Content-Type: application/json" \
  -X POST http://localhost:15672/api/queues/%2F/task_jobs_dlq/get \
  -d '{"count":5,"ackmode":"ack_requeue_true","encoding":"auto","truncate":50000}' | python3 -m json.tool
```

---

## 8. Проверка идемпотентности

Сервис `tasks` всегда генерирует новый `message_id`, поэтому для проверки дубля удобно вручную опубликовать два одинаковых сообщения через RabbitMQ HTTP API.

Опубликовать первое сообщение:

```bash
curl -i -u guest:guest \
  -H "Content-Type: application/json" \
  -X POST http://localhost:15672/api/exchanges/%2F/amq.default/publish \
  -d '{"properties":{"content_type":"application/json","delivery_mode":2,"message_id":"manual-duplicate-001"},"routing_key":"task_jobs","payload":"{\"job\":\"process_task\",\"task_id\":\"t_duplicate\",\"attempt\":1,\"message_id\":\"manual-duplicate-001\"}","payload_encoding":"string"}'
```

Опубликовать второе сообщение с тем же `message_id`:

```bash
curl -i -u guest:guest \
  -H "Content-Type: application/json" \
  -X POST http://localhost:15672/api/exchanges/%2F/amq.default/publish \
  -d '{"properties":{"content_type":"application/json","delivery_mode":2,"message_id":"manual-duplicate-001"},"routing_key":"task_jobs","payload":"{\"job\":\"process_task\",\"task_id\":\"t_duplicate\",\"attempt\":1,\"message_id\":\"manual-duplicate-001\"}","payload_encoding":"string"}'
```

Ожидаемый лог worker:

```text
processed successfully task_id=t_duplicate message_id=manual-duplicate-001; ack
duplicate message_id=manual-duplicate-001 detected; ack without processing
```

Это демонстрирует, что повторная доставка сообщения с тем же `message_id` не приводит к повторному выполнению работы.

---

## 9. Остановка проекта

Остановить `tasks`:

```text
Control + C
```

Остановить `worker`:

```text
Control + C
```

Остановить RabbitMQ без удаления данных:

```bash
cd ~/projects/pz14-task-queue/deploy/rabbit
```

```bash
docker compose down
```

Остановить RabbitMQ с удалением volume и всех очередей:

```bash
cd ~/projects/pz14-task-queue/deploy/rabbit
```

```bash
docker compose down -v
```

---
## 10. Скриншоты
![Скриншот 1](./assets/scr_1.png)
----
![Скриншот 2](./assets/scr_2.png)
----
![Скриншот 3](./assets/scr_3.png)
----
![Скриншот 4](./assets/scr_4.png)
----
![Скриншот 5](./assets/scr_5.png)
----
![Скриншот 6](./assets/scr_6.png)
----
![Скриншот 7](./assets/scr_7.png)

## 11. Контрольные вопросы и ответы

### 1. Чем задача в очереди отличается от простого события?

Событие сообщает, что что-то произошло. Задача описывает работу, которую нужно выполнить. Событие обычно короткое и не требует длительной обработки, а задача может выполняться долго, падать с ошибкой и требовать повторной попытки.

### 2. Зачем нужны retries?

Retries нужны для временных ошибок: кратковременная недоступность внешнего сервиса, сбой соединения, занятый ресурс. Повторная попытка даёт задаче шанс успешно завершиться без ручного вмешательства.

### 3. Почему нельзя бесконечно возвращать ошибочное сообщение в основную очередь?

Бесконечный возврат создаёт вечный цикл обработки. Worker будет снова и снова брать одно и то же проблемное сообщение, тратить ресурсы и мешать обработке нормальных задач.

### 4. Что такое DLQ и зачем она используется?

DLQ, или Dead Letter Queue, — это очередь проблемных сообщений. В неё отправляют сообщения, которые не удалось обработать после допустимого числа попыток. DLQ помогает не потерять сообщение и не блокировать основную очередь.

### 5. Почему в системах очередей возможна повторная доставка одного и того же сообщения?

Очереди часто работают по модели at-least-once delivery. Например, worker уже выполнил работу, но упал до отправки `ack`. RabbitMQ не получил подтверждение и может доставить это же сообщение снова.

### 6. Что такое идемпотентность обработчика?

Идемпотентность означает, что повторная обработка одного и того же сообщения не приводит к повторному выполнению бизнес-операции. Обработчик распознаёт дубль и безопасно подтверждает его без повторной работы.

### 7. Зачем нужен message_id?

`message_id` нужен как уникальный идентификатор сообщения. Worker использует его, чтобы понять, обрабатывалось ли это сообщение раньше.

### 8. Почему хранение обработанных message_id даже в памяти полезно для учебного примера?

Память процесса позволяет просто показать сам принцип идемпотентности без базы данных. Это не промышленное решение, но для практики достаточно, чтобы увидеть, как повторное сообщение пропускается.

### 9. Что произойдёт, если worker выполнит обработку, но не успеет отправить ack?

RabbitMQ будет считать сообщение неподтверждённым. После потери соединения или падения worker сообщение может быть доставлено повторно. Именно поэтому нужна идемпотентность.

### 10. Почему модель producer–consumer удобна для тяжёлых фоновых задач?

Producer быстро принимает запрос и ставит задачу в очередь, а consumer выполняет тяжёлую работу отдельно. Клиент не ждёт долгую операцию, нагрузку можно распределять между несколькими worker, а ошибки можно обрабатывать через retries и DLQ.
