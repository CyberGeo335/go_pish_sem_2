FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/tasks ./cmd/tasks

FROM alpine:3.20

RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=build /out/tasks /app/tasks

USER appuser
EXPOSE 8082

ENV TASKS_PORT=8082 \
    AUTH_BASE_URL=http://auth:8081 \
    LOG_LEVEL=info

ENTRYPOINT ["/app/tasks"]
