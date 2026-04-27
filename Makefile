-include .env
export

.PHONY: build run test clean setup migrate-up migrate-down seed-dev docker-build docker-run docker-stop postman-push \
        build-ApiFunction sam-build sam-deploy sam-migrate sam-logs sam-delete

BINARY=bin/api
MAIN=cmd/api/main.go
IMAGE=weddo-api
VERSION?=latest
SAM_STACK?=weddo-api-prod

build:
	go build -o $(BINARY) $(MAIN)

run:
	go run $(MAIN)

test:
	go test ./... -v

clean:
	rm -rf bin/

setup:
	go mod tidy
	@test -f .env || cp .env.example .env
	@echo "Setup concluído. Edite o .env se necessário."

migrate-up:
	go run $(MAIN) -migrate-up

migrate-down:
	go run $(MAIN) -migrate-down

seed-dev:
	go run $(MAIN) -seed-dev

docker-build:
	docker build -t $(IMAGE):$(VERSION) .

docker-run:
	docker run -d --name $(IMAGE) \
		--env-file .env \
		-p 8080:8080 \
		$(IMAGE):$(VERSION)

docker-stop:
	docker stop $(IMAGE) && docker rm $(IMAGE)

postman-push:
	@test -n "$(POSTMAN_API_KEY)" || (echo "Erro: POSTMAN_API_KEY não definida. Preencha no .env ou exporte." && exit 1)
	cd postman && postman login --with-api-key "$(POSTMAN_API_KEY)" && postman workspace push --yes

# ─────────────────────────── SAM / Lambda ──────────────────────────────────────

# Chamado automaticamente pelo `sam build` (BuildMethod: makefile no template.yaml).
# Compila o binário para arm64 Linux e copia as migrations para o pacote Lambda.
build-ApiFunction:
	GOARCH=arm64 GOOS=linux CGO_ENABLED=0 \
		go build -tags lambda.norpc -ldflags="-s -w" \
		-o $(ARTIFACTS_DIR)/bootstrap ./cmd/lambda/
	cp -r migrations/ $(ARTIFACTS_DIR)/migrations/

# Compila todos os recursos SAM (usa o target build-ApiFunction acima).
sam-build:
	sam build

# Faz o deploy na AWS. Lê variáveis do .env (Payment/CORS/Google opcionais — vazio ok).
sam-deploy: sam-build
	@test -n "$(DATABASE_URL)" || (echo "Erro: DATABASE_URL não definida. Adicione ao .env." && exit 1)
	@test -n "$(JWT_SECRET)"   || (echo "Erro: JWT_SECRET não definida. Adicione ao .env." && exit 1)
	sam deploy --config-env prod \
		--parameter-overrides "DatabaseURL=$(DATABASE_URL) JWTSecret=$(JWT_SECRET) PaymentProvider=$(PAYMENT_PROVIDER) IPHandle=$(IP_HANDLE) IPRedirectURL=$(IP_REDIRECT_URL) IPWebhookURL=$(IP_WEBHOOK_URL) CORSAllowedOrigins=$(CORS_ALLOWED_ORIGINS) GoogleOAuthClientId=$(GOOGLE_OAUTH_CLIENT_ID) GoogleOAuthRedirectURL=$(GOOGLE_OAUTH_REDIRECT_URL) GoogleOAuthClientSecret=$(GOOGLE_OAUTH_CLIENT_SECRET) GoogleOAuthTokenCipherKey=$(GOOGLE_OAUTH_TOKEN_CIPHER_KEY)"

# ─── Migrações ────────────────────────────────────────────────────────────────
# Fluxo: deploy com RUN_MIGRATIONS=true → invoca a Lambda via health check → reverte.
sam-migrate:
	@echo ">>> Ativando RUN_MIGRATIONS=true e fazendo deploy..."
	sam deploy --config-env prod \
		--parameter-overrides "DatabaseURL=$(DATABASE_URL) JWTSecret=$(JWT_SECRET) RunMigrations=true"
	@echo ">>> Invocando Lambda para disparar migração (cold start executa migrate Up)..."
	sam remote invoke ApiFunction --stack-name $(SAM_STACK) \
		--event '{"version":"2.0","routeKey":"GET /api/v1/health","rawPath":"/api/v1/health","rawQueryString":"","headers":{"host":"localhost"},"requestContext":{"accountId":"","apiId":"","domainName":"","domainPrefix":"","http":{"method":"GET","path":"/api/v1/health","protocol":"HTTP/1.1","sourceIp":"","userAgent":""},"requestId":"","routeKey":"GET /api/v1/health","stage":"$$default","time":"","timeEpoch":0}}'
	@echo ""
	@echo ">>> Revertendo RUN_MIGRATIONS=false..."
	sam deploy --config-env prod \
		--parameter-overrides "DatabaseURL=$(DATABASE_URL) JWTSecret=$(JWT_SECRET) RunMigrations=false"
	@echo ">>> Migrações concluídas."

# Tail dos logs da Lambda em tempo real.
sam-logs:
	sam logs -n ApiFunction --stack-name $(SAM_STACK) --tail

# Remove completamente o stack (atenção: o banco fará snapshot antes de deletar).
sam-delete:
	sam delete --stack-name $(SAM_STACK)
