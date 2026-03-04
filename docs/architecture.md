# Arquitetura

## Clean Architecture

O projeto segue Clean Architecture com dependências apontando para o centro (domínio).

```
┌─────────────────────────────────────────┐
│  infra/web (handlers, middlewares)      │
│  infra/database (repositórios SQLite)   │
│  infra/gateway (Mercado Pago)           │
│  infra/config                           │
│  ┌───────────────────────────────────┐  │
│  │  usecase (regras de aplicação)    │  │
│  │  dto (request/response)           │  │
│  │  ┌─────────────────────────────┐  │  │
│  │  │  domain (entidades,         │  │  │
│  │  │  interfaces de repositório) │  │  │
│  │  └─────────────────────────────┘  │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

### Regra de dependência

- **domain/** não importa nenhum outro pacote interno.
- **usecase/** importa apenas `domain/`.
- **dto/** não importa nenhum outro pacote interno (tipos puros de transporte).
- **infra/** importa `domain/`, `usecase/` e `dto/`.

### Camadas

| Camada | Pacote | Responsabilidade |
|--------|--------|------------------|
| Domain | `internal/domain/entity` | Entidades com regras de negócio intrínsecas |
| Domain | `internal/domain/repository` | Interfaces dos repositórios (contratos) |
| Use Case | `internal/usecase/*` | Orquestração de regras de negócio por contexto |
| DTO | `internal/dto` | Structs de request/response para a camada HTTP |
| Infra | `internal/infra/database` | Implementação concreta dos repositórios |
| Infra | `internal/infra/gateway` | Clientes de serviços externos (Mercado Pago) |
| Infra | `internal/infra/web/handler` | Handlers HTTP |
| Infra | `internal/infra/web/middleware` | Middlewares (auth, CORS, logging, recovery, tenant) |
| Infra | `internal/infra/config` | Leitura de variáveis de ambiente |
| Entrypoint | `cmd/api` | Bootstrap: config → DB → repos → use cases → handlers → router → server |

## Multi-tenancy

### Estratégia: shared schema com discriminador

Banco único, todas as tabelas possuem `wedding_id`. Simples, eficiente, e suficiente para a escala esperada. Se necessário no futuro, migrar para PostgreSQL com schema-per-tenant ou database-per-tenant.

### Resolução do tenant

| Contexto | Como resolve | Onde |
|----------|-------------|------|
| Endpoints públicos | Slug na URL: `/api/v1/w/{slug}/...` | Middleware `TenantResolver` |
| Endpoints admin | `wedding_id` no JWT claims | Middleware `Auth` |
| Webhook de pagamento | `payment → gift → wedding` no banco | Handler direto |

### Fluxo de uma request pública

```
GET /api/v1/w/manu-rafa/gifts
  → TenantResolver middleware
    → Busca wedding por slug
    → Injeta wedding_id no context
  → Handler
    → Use Case (recebe wedding_id via context)
      → Repository (filtra por wedding_id)
```

### Fluxo de uma request admin

```
GET /api/v1/admin/guests
  → Auth middleware
    → Valida JWT
    → Extrai wedding_id dos claims
    → Injeta no context
  → Handler
    → Use Case (recebe wedding_id via context)
      → Repository (filtra por wedding_id)
```

### Isolamento

- Repositórios **sempre** recebem `weddingID` como parâmetro.
- Nunca existe query sem filtro de tenant (exceto busca de wedding por slug/email).
- Use cases não decidem o tenant — recebem do handler via context.

## Decisões Técnicas

### Go 1.23+

Desempenho, simplicidade, tipagem estática e excelente stdlib para HTTP.

### chi (router)

Compatível com `net/http`, middleware chain, agrupamento de rotas, parâmetros de URL. Leve e idiomático.

### SQLite

Banco embutido, zero configuração. Ideal para o escopo inicial. Multi-tenancy via coluna `wedding_id` em todas as tabelas.

### golang-migrate

Migrações versionadas em SQL puro (up/down), executáveis no boot da aplicação.

### JWT para autenticação admin

Token stateless com `wedding_id` nos claims. Cada casamento tem seu próprio admin (email + senha). O login gera um token JWT assinado com segredo global.

### Mercado Pago (pagamentos)

Gateway para lista de presentes. SDK oficial em Go. PIX (~0.5% taxa) e cartão (~4-5%). Checkout Transparente. Inicialmente com credenciais globais (via env). Detalhes em [gift-list.md](gift-list.md).

### Validação com go-playground/validator

Validação declarativa via struct tags. Erros traduzidos para respostas HTTP padronizadas.

### slog (stdlib)

Logger estruturado nativo do Go 1.21+. JSON em produção, texto em desenvolvimento.

## CORS

Configurável via env. Em multi-tenant, aceitar origens de diferentes frontends:

```
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://manurafa.com.br
```

## Estrutura de Diretórios

```
mr-wedding-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── wedding.go
│   │   │   ├── invitation.go
│   │   │   ├── guest.go
│   │   │   ├── gift.go
│   │   │   └── payment.go
│   │   └── repository/
│   │       ├── wedding.go
│   │       ├── invitation.go
│   │       ├── guest.go
│   │       ├── gift.go
│   │       └── payment.go
│   ├── usecase/
│   │   ├── wedding/
│   │   │   └── wedding.go
│   │   ├── rsvp/
│   │   │   └── rsvp.go
│   │   ├── guest/
│   │   │   └── guest.go
│   │   ├── invitation/
│   │   │   └── invitation.go
│   │   ├── gift/
│   │   │   └── gift.go
│   │   └── payment/
│   │       └── payment.go
│   ├── dto/
│   │   ├── request.go
│   │   └── response.go
│   └── infra/
│       ├── config/
│       │   └── config.go
│       ├── database/
│       │   ├── sqlite.go
│       │   ├── wedding_repository.go
│       │   ├── invitation_repository.go
│       │   ├── guest_repository.go
│       │   ├── gift_repository.go
│       │   └── payment_repository.go
│       ├── gateway/
│       │   └── mercadopago.go
│       └── web/
│           ├── handler/
│           │   ├── wedding.go
│           │   ├── rsvp.go
│           │   ├── gift.go
│           │   ├── payment.go
│           │   ├── guest.go
│           │   ├── invitation.go
│           │   ├── auth.go
│           │   └── dashboard.go
│           ├── middleware/
│           │   ├── auth.go
│           │   ├── tenant.go
│           │   └── logger.go
│           └── router.go
├── migrations/
│   ├── 001_create_weddings.up.sql
│   ├── 001_create_weddings.down.sql
│   ├── 002_create_invitations.up.sql
│   ├── 002_create_invitations.down.sql
│   ├── 003_create_guests.up.sql
│   ├── 003_create_guests.down.sql
│   ├── 004_create_gifts.up.sql
│   ├── 004_create_gifts.down.sql
│   ├── 005_create_payments.up.sql
│   └── 005_create_payments.down.sql
├── docs/
├── .cursor/rules/
├── .env.example
├── .gitignore
├── Makefile
├── Dockerfile
├── go.mod
└── go.sum
```
