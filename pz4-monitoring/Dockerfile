FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/pz4-monitoring ./cmd/server

FROM alpine:3.20

RUN adduser -D -g '' appuser
USER appuser

COPY --from=builder /bin/pz4-monitoring /bin/pz4-monitoring

EXPOSE 8080
CMD ["/bin/pz4-monitoring"]
