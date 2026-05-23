# Практическое занятие №16 — публикация Go-приложения в Kubernetes

Проект содержит контейнеризированный backend-сервис `tasks` на Go и минимальные Kubernetes-манифесты для публикации приложения через `ConfigMap`, `Deployment`, `Service`, readiness probe, liveness probe и `kubectl port-forward`.

## Что реализовано

- HTTP backend-сервис `tasks` на Go без внешних зависимостей.
- Endpoint проверки состояния: `GET /health`.
- Простые in-memory CRUD endpoints для задач.
- `Dockerfile` для сборки образа `techip-tasks:0.1`.
- Kubernetes-манифесты в `deploy/k8s/`:
  - `configmap.yaml`;
  - `deployment.yaml`;
  - `service.yaml`.
- Readiness и liveness probes на `/health`.
- Команды для запуска на macOS через Docker, `kubectl` и `kind`.
- Ответы на контрольные вопросы в `docs/control-questions.md`.

## Структура проекта

```text
pz16-tasks-k8s/
├── cmd/
│   └── tasks/
│       ├── main.go
│       └── main_test.go
├── deploy/
│   └── k8s/
│       ├── configmap.yaml
│       ├── deployment.yaml
│       └── service.yaml
├── docs/
│   └── control-questions.md
├── .dockerignore
├── Dockerfile
├── go.mod
└── README.md
```

## API приложения

| Метод | Путь | Назначение |
|---|---|---|
| `GET` | `/` | Информация о сервисе |
| `GET` | `/health` | Health check для Kubernetes probes |
| `GET` | `/tasks` | Получить список задач |
| `POST` | `/tasks` | Создать задачу |
| `GET` | `/tasks/{id}` | Получить задачу по ID |
| `PATCH` | `/tasks/{id}` | Обновить задачу |
| `DELETE` | `/tasks/{id}` | Удалить задачу |

## Переменные окружения

| Переменная | Значение по умолчанию | Назначение |
|---|---:|---|
| `TASKS_PORT` | `8082` | HTTP-порт сервиса |
| `AUTH_BASE_URL` | `http://auth:8081` | Адрес внешнего сервиса авторизации для демонстрации конфигурации |
| `LOG_LEVEL` | `info` | Уровень логирования |

В Kubernetes эти значения передаются через `ConfigMap`.

---

# Команды от А до Я для macOS

Ниже приведён полный порядок выполнения. Команды не вынесены в скрипты специально: их нужно выполнить вручную и по отдельности.

## 1. Установить инструменты

Установить kubectl:

```bash
brew install go kubectl kind
```

Проверьте версии инструментов:

```bash
go version
```

```bash
docker version
```

```bash
kubectl version --client
```

```bash
kind version
```


## 2. Проверить Go-приложение локально

```bash
go test ./...
```

Запустите сервис:

```bash
go run ./cmd/tasks
```

Откройте второй терминал в этой же папке и проверьте `/health`:

```bash
curl -i http://localhost:8082/health
```

Проверьте список задач:

```bash
curl -i http://localhost:8082/tasks
```

Создайте задачу:

```bash
curl -i -X POST http://localhost:8082/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Проверить Kubernetes Deployment","done":false}'
```

Остановите локальный запуск в первом терминале сочетанием клавиш:

```text
Ctrl + C
```

## 5. Собрать Docker-образ с фиксированным тегом

```bash
docker build -t techip-tasks:0.1 .
```

Проверьте, что образ создан:

```bash
docker images | grep techip-tasks
```

## 6. Проверить Docker-образ локально

```bash
docker run --rm -p 8082:8082 \
  -e TASKS_PORT=8082 \
  -e AUTH_BASE_URL=http://auth:8081 \
  -e LOG_LEVEL=info \
  techip-tasks:0.1
```

Во втором терминале проверьте health endpoint:

```bash
curl -i http://localhost:8082/health
```

Остановите контейнер в первом терминале:

```text
Ctrl + C
```

## 7. Создать локальный Kubernetes-кластер через kind

```bash
kind create cluster --name pz16
```

Проверьте доступ к кластеру:

```bash
kubectl cluster-info --context kind-pz16
```

```bash
kubectl get nodes
```

## 8. Загрузить Docker-образ внутрь kind-кластера

В этой работе образ становится доступен Kubernetes-кластеру через загрузку локального Docker-образа в `kind`:

```bash
kind load docker-image techip-tasks:0.1 --name pz16
```

Это важно: без этой команды `kind` может не увидеть локальный образ, даже если он есть в Docker Desktop.

## 9. Проверить Kubernetes-манифесты

```bash
cat deploy/k8s/configmap.yaml
```

```bash
cat deploy/k8s/deployment.yaml
```

```bash
cat deploy/k8s/service.yaml
```

## 10. Применить ConfigMap

```bash
kubectl apply -f deploy/k8s/configmap.yaml
```

## 11. Применить Deployment

```bash
kubectl apply -f deploy/k8s/deployment.yaml
```

## 12. Применить Service

```bash
kubectl apply -f deploy/k8s/service.yaml
```

## 13. Проверить Pod

```bash
kubectl get pods
```

Сохраните имя Pod в переменную:

```bash
POD_NAME=$(kubectl get pods -l app=tasks -o jsonpath='{.items[0].metadata.name}')
```

Проверьте подробное состояние Pod:

```bash
kubectl describe pod "$POD_NAME"
```

## 14. Проверить Deployment

```bash
kubectl get deployment
```

```bash
kubectl describe deployment tasks
```

## 15. Проверить Service

```bash
kubectl get svc
```

```bash
kubectl describe svc tasks
```

## 16. Посмотреть логи контейнера

```bash
kubectl logs "$POD_NAME"
```

## 17. Проверить доступ через port-forward

В первом терминале выполните:

```bash
kubectl port-forward svc/tasks 8082:8082
```

Во втором терминале проверьте endpoint `/health`:

```bash
curl -i http://localhost:8082/health
```

Проверьте API задач:

```bash
curl -i http://localhost:8082/tasks
```

Создайте задачу через сервис в Kubernetes:

```bash
curl -i -X POST http://localhost:8082/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Задача создана через Kubernetes Service","done":false}'
```

Остановите `port-forward` в первом терминале:

```text
Ctrl + C
```

## 20. Удалить Kubernetes-ресурсы после проверки

Удалять ресурсы нужно в обратном порядке:

```bash
kubectl delete -f deploy/k8s/service.yaml
```

```bash
kubectl delete -f deploy/k8s/deployment.yaml
```

```bash
kubectl delete -f deploy/k8s/configmap.yaml
```

## 21. Удалить kind-кластер, если он больше не нужен

```bash
kind delete cluster --name pz16
```
---

# Контрольные вопросы

Ответы вынесены в отдельный файл:

```text
docs/control-questions.md
```


# Контрольные вопросы

## 1. Что такое Kubernetes и для чего он используется?

Kubernetes — это система оркестрации контейнеров. Она используется для запуска, масштабирования, обновления и сопровождения контейнеризированных приложений.

## 2. Чем Pod отличается от Deployment?

Pod — минимальная единица запуска контейнера в Kubernetes. Deployment — объект более высокого уровня, который управляет Pod, поддерживает нужное количество реплик и пересоздаёт Pod при сбоях.

## 3. Почему приложение в Kubernetes обычно публикуют через Deployment, а не через одиночный Pod?

Одиночный Pod неудобен для сопровождения: если он завершится или будет удалён, его нужно создавать заново вручную. Deployment описывает желаемое состояние и автоматически поддерживает нужное число Pod.

## 4. Зачем нужен Service и почему нельзя строить обращение к приложению напрямую через Pod?

Pod имеет нестабильный жизненный цикл и IP-адрес. Service создаёт стабильную точку доступа и направляет трафик на подходящие Pod по selector.

## 5. Что такое ConfigMap?

ConfigMap — объект Kubernetes для хранения несекретной конфигурации приложения: портов, адресов сервисов, уровней логирования и других параметров окружения.

## 6. Чем ConfigMap отличается от Secret?

ConfigMap предназначен для обычной несекретной конфигурации. Secret используется для чувствительных данных: паролей, токенов, ключей доступа.

## 7. Для чего используется readiness probe?

Readiness probe показывает, готов ли контейнер принимать трафик. Пока проверка не проходит, Service не направляет запросы на Pod.

## 8. Для чего используется liveness probe?

Liveness probe показывает, живо ли приложение. Если проверка стабильно падает, Kubernetes считает контейнер неисправным и перезапускает его.

## 9. Почему важно использовать фиксированный тег образа, а не только latest?

Фиксированный тег делает развёртывание воспроизводимым. По нему понятно, какая версия приложения запущена в кластере. Тег latest может указывать на разные версии в разное время.

## 10. Зачем нужен kubectl port-forward?

kubectl port-forward пробрасывает локальный порт компьютера на Service или Pod в Kubernetes. Это позволяет проверить приложение из браузера или через curl без Ingress и внешнего LoadBalancer.

## 11. Что делает команда kubectl scale deployment ...?

Команда изменяет количество реплик Deployment. Kubernetes создаёт или удаляет Pod, чтобы привести фактическое состояние к указанному числу реплик.

## 12. Почему публикация приложения в Kubernetes считается декларативной?

Разработчик описывает желаемое состояние в YAML-манифестах: какие ресурсы нужны, какой образ запускать, сколько реплик поддерживать. Kubernetes сам приводит кластер к этому состоянию.
