.PHONY: certs db-up db-down tidy run curl-health curl-student curl-email curl-injection-test git-init

certs:
	mkdir -p certs
	openssl req -x509 -newkey rsa:2048 -nodes \
		-keyout certs/server.key \
		-out certs/server.crt \
		-days 365 \
		-subj "/CN=localhost" \
		-addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

tidy:
	go mod tidy

run:
	go run ./cmd/server

curl-health:
	curl -k https://localhost:8443/health

curl-student:
	curl -k "https://localhost:8443/students?id=1"

curl-email:
	curl -k "https://localhost:8443/students/by-email?email=ivanov@example.com"

curl-injection-test:
	curl -k "https://localhost:8443/students?id=1%20OR%201=1"

git-init:
	git init
	git add .
	git commit -m "Initial Go HTTPS SQL injection protection practice"
	git branch -M main
	git remote add origin https://github.com/CyberGeo335/pz5-security.git
