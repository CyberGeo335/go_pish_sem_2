# Практическое занятие №2 — gRPC на Go


В работе реализован учебный gRPC-микросервис `StudentService` на Go. Базовая часть содержит методы `Ping` и `GetStudentByID`. Дополнительно реализованы задания из конца практики: список студентов, создание студента и поле `specialization`.

## Что реализовано

- `.proto`-контракт сервиса;
- генерация Go-кода через `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`;
- gRPC-сервер на порту `50051`;
- gRPC-клиент, который вызывает методы сервиса;
- обработка ошибки `NotFound` при запросе несуществующего студента;
- дополнительные методы `ListStudents` и `CreateStudent`;
- поле `specialization` в структуре `Student`;
- краткое сравнение с REST API;
- ответы на контрольные вопросы.

## Структура проекта

```text
pz2-grpc/
  proto/
    student.proto
  cmd/
    server/
      main.go
    client/
      main.go
  internal/
    student/
      data.go
      service.go
  gen/
    studentpb/
      generated files after make proto
  scripts/
    bootstrap_mac.sh
  go.mod
  Makefile
  README.md
```

## Требования для macOS

Проверьте Go:

```bash
go version
```

Если `protoc` не установлен, установите через Homebrew:

```bash
brew install protobuf
protoc --version
```

Установите Go-плагины генерации:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Быстрый запуск на macOS

Из корня проекта:

```bash
chmod +x scripts/bootstrap_mac.sh
./scripts/bootstrap_mac.sh
```

Скрипт установит Go-инструменты генерации, подтянет зависимости, сгенерирует protobuf/gRPC-код и выполнит `go mod tidy`.

## Ручной запуск по шагам

### 1. Инициализация зависимостей

```bash
go get google.golang.org/grpc
go get google.golang.org/protobuf
go mod tidy
```

### 2. Запуск сервера

В первом терминале:

```bash
make server
```

Ожидаемый вывод:

```text
gRPC server started on :50051
```

### 4. Запуск клиента

Во втором терминале:

```bash
make client
```

Пример ожидаемого вывода:

```text
Ping response: Server received: hello grpc
Student by id=1: id=1, full_name=Иванов Иван Иванович, group=ИВБО-01-25, email=ivanov@example.com, specialization=Backend-разработка
Students count: 3
Student from list: id=1, full_name=Иванов Иван Иванович, group=ИВБО-01-25, email=ivanov@example.com, specialization=Backend-разработка
Student from list: id=2, full_name=Петрова Мария Сергеевна, group=ИВБО-02-25, email=petrova@example.com, specialization=DevOps
Student from list: id=3, full_name=Сидоров Алексей Андреевич, group=ИВБО-03-25, email=sidorov@example.com, specialization=Информационная безопасность
Created student: id=4, full_name=Смирнов Дмитрий Олегович, group=ИВБО-04-25, email=smirnov@example.com, specialization=Golang и микросервисы
Expected error for id=999: NotFound student not found
```

## 3. Скриншоты
![Скриншот 1](./assets/Снимок%20экрана%C2%A0—%202026-05-03%20в%C2%A013.03.43.png)
----
![Скриншот 2](./assets/Снимок%20экрана%C2%A0—%202026-05-03%20в%C2%A013.04.01.png)
----
![Скриншот 3](./assets/Снимок%20экрана%C2%A0—%202026-05-03%20в%C2%A013.04.26.png)
----

## Как тот же сервис выглядел бы в REST

В REST API сервис можно было бы представить так:

| Действие | Метод и URL | Тело запроса | Ответ |
|---|---|---|---|
| Проверка сервера | `GET /ping?message=hello` | нет | `{ "message": "Server received: hello" }` |
| Получить студента | `GET /students/1` | нет | `{ "student": { "id": 1, "full_name": "...", "group": "...", "email": "...", "specialization": "..." } }` |
| Получить список | `GET /students` | нет | `{ "students": [...] }` |
| Создать студента | `POST /students` | `{ "full_name": "...", "group": "...", "email": "...", "specialization": "..." }` | `{ "student": { ... } }` |

Главное отличие: в REST вручную проектируются URL и JSON-формат, а в gRPC сначала описывается строгий контракт в `.proto`, затем по нему генерируется типизированный клиент и серверный интерфейс.

## Контрольные вопросы

### 1. Что такое gRPC?

gRPC — это фреймворк удаленного вызова процедур. Он позволяет одному приложению вызывать методы другого приложения так, как будто это обычные локальные методы.

### 2. Какую роль играет `.proto`-файл?

`.proto`-файл описывает контракт сервиса: сообщения, поля, сервисы и методы. По нему генерируется код для клиента и сервера.

### 3. Для чего нужен `protoc`?

`protoc` — это компилятор Protocol Buffers. Он читает `.proto`-файл и запускает плагины генерации кода для нужного языка.

### 4. Зачем используются `protoc-gen-go` и `protoc-gen-go-grpc`?

`protoc-gen-go` генерирует Go-структуры для protobuf-сообщений. `protoc-gen-go-grpc` генерирует Go-интерфейсы и клиент/серверный код для gRPC-сервисов.

### 5. Чем gRPC отличается от HTTP JSON API?

В HTTP JSON API разработчик обычно вручную описывает URL, JSON-запросы и JSON-ответы. В gRPC контракт описывается в `.proto`, а клиентский и серверный код генерируется автоматически.

### 6. Почему контракт в gRPC считается строго типизированным?

Потому что типы полей и методы заранее описаны в `.proto`. Если клиент или сервер нарушает контракт, ошибка обычно проявляется уже на этапе компиляции или вызова метода.

### 7. Что делает gRPC-клиент в этой работе?

Клиент подключается к серверу на `localhost:50051`, вызывает методы `Ping`, `GetStudentByID`, `ListStudents`, `CreateStudent` и проверяет ошибку при запросе студента с ID `999`.

### 8. Что происходит, если клиент запрашивает несуществующего студента?

Сервер возвращает gRPC-ошибку со статусом `NotFound` и сообщением `student not found`.

### 9. Почему для локальной учебной среды допустимо использовать insecure credentials?

Потому что клиент и сервер запускаются локально на одной машине. Для реальной системы нужно использовать защищенное соединение, например TLS.

### 10. В каких случаях gRPC особенно удобен в backend-разработке?

gRPC удобен для внутреннего взаимодействия микросервисов, когда нужен строгий контракт, типизированные сообщения, компактная сериализация и автоматически сгенерированный клиент.