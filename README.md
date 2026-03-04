# mr-wedding-api

API multi-tenant para sites de casamento. Fornece confirmação de presença (RSVP), lista de presentes com pagamento integrado (PIX/cartão via Mercado Pago) e painel administrativo.

Cada casamento é um tenant isolado. O primeiro tenant é o casamento **Manoela & Rafael — 07.07.2026** ([frontend](../mr-wedding/)).

## Stack

| Componente | Tecnologia |
|------------|------------|
| Linguagem | Go 1.23+ |
| Arquitetura | Clean Architecture, multi-tenant |
| Router HTTP | [chi](https://github.com/go-chi/chi) |
| Banco de dados | SQLite (via [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)) |
| Migrações | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Configuração | Variáveis de ambiente (envconfig) |
| Autenticação admin | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) |
| Validação | [go-playground/validator](https://github.com/go-playground/validator) |
| Pagamentos | [Mercado Pago SDK Go](https://github.com/mercadopago/sdk-go) (PIX + cartão) |
| Logging | `log/slog` (stdlib) |
| Testes | `testing` (stdlib) + [testify](https://github.com/stretchr/testify) |

## Quick Start

```bash
cp .env.example .env
make setup
make run
```

O servidor sobe em `http://localhost:8080`.

## Estrutura do Projeto

```
├── cmd/api/              # Entrypoint
├── internal/
│   ├── domain/           # Entidades e interfaces de repositório
│   ├── usecase/          # Casos de uso (regras de negócio)
│   ├── dto/              # Objetos de transferência (request/response)
│   └── infra/
│       ├── database/     # Repositórios SQLite
│       ├── gateway/      # Clientes externos (Mercado Pago)
│       ├── web/          # Handlers, middlewares, router
│       └── config/       # Configuração
├── migrations/           # SQL migrations
├── docs/                 # Documentação detalhada
└── .cursor/rules/        # Convenções para o Cursor AI
```

## Documentação

| Documento | Conteúdo |
|-----------|----------|
| [docs/roadmap.md](docs/roadmap.md) | Roadmap por fases e prioridades |
| [docs/architecture.md](docs/architecture.md) | Arquitetura, multi-tenancy e decisões técnicas |
| [docs/api.md](docs/api.md) | Endpoints, contratos e exemplos |
| [docs/database.md](docs/database.md) | Modelo de dados e migrações |
| [docs/gift-list.md](docs/gift-list.md) | Lista de presentes: fluxos e integração Mercado Pago |
