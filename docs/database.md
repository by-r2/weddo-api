# Modelo de Dados

## Multi-tenancy

Todas as tabelas de negócio possuem `wedding_id` como chave estrangeira para `weddings`. Isso garante isolamento de dados entre casamentos. Queries sempre filtram por `wedding_id` — nunca acessam dados de outro tenant.

## Diagrama

```
┌──────────────────────────────┐
│          weddings            │
├──────────────────────────────┤
│ id               TEXT PK     │──────────────────────────────────┐
│ slug             TEXT UNIQUE │                                  │
│ title            TEXT        │                                  │
│ date             DATE        │                                  │
│ partner1_name    TEXT        │                                  │
│ partner2_name    TEXT        │                                  │
│ admin_email      TEXT        │                                  │
│ admin_pass_hash  TEXT        │                                  │
│ active           INTEGER     │                                  │
│ created_at       DATETIME    │                                  │
│ updated_at       DATETIME    │                                  │
└──────────────────────────────┘                                  │
         │                                                        │
         │  wedding_id FK                                         │
         ▼                                                        │
┌──────────────────────────┐       ┌──────────────────────────────┐
│       invitations        │       │           guests             │
├──────────────────────────┤       ├──────────────────────────────┤
│ id          TEXT PK      │◄──┐   │ id             TEXT PK       │
│ wedding_id  TEXT FK      │   │   │ invitation_id  TEXT FK       │
│ code        TEXT         │   └───│ wedding_id     TEXT FK       │
│ label       TEXT         │       │ name           TEXT          │
│ max_guests  INTEGER      │       │ phone          TEXT          │
│ notes       TEXT         │       │ email          TEXT          │
│ created_at  DATETIME     │       │ status         TEXT          │
│ updated_at  DATETIME     │       │ confirmed_at   DATETIME     │
└──────────────────────────┘       │ created_at     DATETIME     │
                                   │ updated_at     DATETIME     │
         │  wedding_id FK          └──────────────────────────────┘
         ▼
┌──────────────────────────────┐       ┌──────────────────────────────┐
│           gifts              │       │          payments            │
├──────────────────────────────┤       ├──────────────────────────────┤
│ id            TEXT PK        │◄──┐   │ id              TEXT PK      │
│ wedding_id    TEXT FK        │   │   │ gift_id         TEXT FK      │
│ name          TEXT           │   └───│ wedding_id      TEXT FK      │
│ description   TEXT           │       │ provider_id     TEXT         │
│ price         REAL           │       │ amount          REAL         │
│ image_url     TEXT           │       │ status          TEXT         │
│ category      TEXT           │       │ payment_method  TEXT         │
│ status        TEXT           │       │ payer_name      TEXT         │
│ created_at    DATETIME       │       │ payer_email     TEXT         │
│ updated_at    DATETIME       │       │ message         TEXT         │
└──────────────────────────────┘       │ pix_qr_code     TEXT         │
                                       │ pix_expiration  DATETIME     │
                                       │ paid_at         DATETIME     │
                                       │ created_at      DATETIME     │
                                       │ updated_at      DATETIME     │
                                       └──────────────────────────────┘
```

## Tabelas

### weddings

Tenant principal. Cada casamento é uma instância isolada do sistema.

| Coluna | Tipo | Restrições | Descrição |
|--------|------|------------|-----------|
| id | TEXT | PK | UUID v4 |
| slug | TEXT | UNIQUE, NOT NULL | Identificador URL-friendly para uso no frontend (ex: "manoela-rafael") |
| title | TEXT | NOT NULL | Nome do evento (ex: "Casamento Manoela & Rafael") |
| date | DATE | | Data do casamento |
| partner1_name | TEXT | NOT NULL | Nome do(a) primeiro(a) noivo(a) |
| partner2_name | TEXT | NOT NULL | Nome do(a) segundo(a) noivo(a) |
| admin_email | TEXT | NOT NULL, UNIQUE | Email para login admin |
| admin_pass_hash | TEXT | NOT NULL | Hash bcrypt da senha admin |
| active | INTEGER | NOT NULL, DEFAULT 1 | 1 = ativo, 0 = desativado |
| created_at | DATETIME | NOT NULL | Timestamp de criação |
| updated_at | DATETIME | NOT NULL | Timestamp da última atualização |

### invitations

Convite físico enviado a um grupo (família, casal, pessoa individual).

| Coluna | Tipo | Restrições | Descrição |
|--------|------|------------|-----------|
| id | TEXT | PK | UUID v4 |
| wedding_id | TEXT | FK → weddings(id), NOT NULL | Casamento ao qual pertence |
| code | TEXT | NOT NULL | Identificador legível (ex: "SILVA-001") |
| label | TEXT | NOT NULL | Nome do grupo (ex: "Família Silva") |
| max_guests | INTEGER | NOT NULL, DEFAULT 1 | Máximo de convidados neste convite |
| notes | TEXT | | Observações internas |
| created_at | DATETIME | NOT NULL | Timestamp de criação |
| updated_at | DATETIME | NOT NULL | Timestamp da última atualização |

Constraint: `UNIQUE(wedding_id, code)` — código único por casamento.

### guests

Pessoa individual vinculada a um convite.

| Coluna | Tipo | Restrições | Descrição |
|--------|------|------------|-----------|
| id | TEXT | PK | UUID v4 |
| invitation_id | TEXT | FK → invitations(id), NOT NULL | Convite ao qual pertence |
| wedding_id | TEXT | FK → weddings(id), NOT NULL | Casamento (desnormalizado para queries diretas) |
| name | TEXT | NOT NULL | Nome como está no convite |
| phone | TEXT | | Telefone (opcional) |
| email | TEXT | | Email (opcional) |
| status | TEXT | NOT NULL, DEFAULT 'pending' | `pending`, `confirmed` ou `declined` |
| confirmed_at | DATETIME | | Timestamp da confirmação |
| created_at | DATETIME | NOT NULL | Timestamp de criação |
| updated_at | DATETIME | NOT NULL | Timestamp da última atualização |

### gifts

Presente virtual no catálogo. O valor vai para a conta dos noivos.

| Coluna | Tipo | Restrições | Descrição |
|--------|------|------------|-----------|
| id | TEXT | PK | UUID v4 |
| wedding_id | TEXT | FK → weddings(id), NOT NULL | Casamento ao qual pertence |
| name | TEXT | NOT NULL | Nome do presente |
| description | TEXT | | Descrição detalhada |
| price | REAL | NOT NULL | Valor em reais |
| image_url | TEXT | | URL da imagem |
| category | TEXT | NOT NULL | Categoria (ex: "Cozinha", "Lua de Mel") |
| status | TEXT | NOT NULL, DEFAULT 'available' | `available` ou `purchased` |
| created_at | DATETIME | NOT NULL | Timestamp de criação |
| updated_at | DATETIME | NOT NULL | Timestamp da última atualização |

### payments

Transação de pagamento associada a um presente.

| Coluna | Tipo | Restrições | Descrição |
|--------|------|------------|-----------|
| id | TEXT | PK | UUID v4 |
| gift_id | TEXT | FK → gifts(id), NOT NULL | Presente sendo comprado |
| wedding_id | TEXT | FK → weddings(id), NOT NULL | Casamento (desnormalizado) |
| provider_id | TEXT | | ID da transação no provedor (InfinitePay slug ou Mercado Pago ID) |
| amount | REAL | NOT NULL | Valor cobrado em reais |
| status | TEXT | NOT NULL, DEFAULT 'pending' | `pending`, `approved`, `rejected`, `expired` |
| payment_method | TEXT | NOT NULL | `pix` ou `credit_card` |
| payer_name | TEXT | NOT NULL | Nome de quem presenteia |
| payer_email | TEXT | | Email do pagador |
| message | TEXT | | Mensagem para os noivos |
| pix_qr_code | TEXT | | Código copia-e-cola do PIX |
| pix_expiration | DATETIME | | Quando o QR code expira |
| paid_at | DATETIME | | Timestamp do pagamento confirmado |
| created_at | DATETIME | NOT NULL | Timestamp de criação |
| updated_at | DATETIME | NOT NULL | Timestamp da última atualização |

## Enums de Status

### Guest Status

| Valor | Significado |
|-------|-------------|
| `pending` | Ainda não respondeu |
| `confirmed` | Confirmou presença |
| `declined` | Informou que não irá |

### Payment Status

| Valor | Significado |
|-------|-------------|
| `pending` | Aguardando pagamento |
| `approved` | Pagamento confirmado |
| `rejected` | Pagamento recusado |
| `expired` | QR code PIX expirou |

### Gift Status

| Valor | Significado |
|-------|-------------|
| `available` | Disponível para compra |
| `purchased` | Já comprado |

## Migrações

```
migrations/
├── 001_create_weddings.up.sql
├── 001_create_weddings.down.sql
├── 002_create_invitations.up.sql
├── 002_create_invitations.down.sql
├── 003_create_guests.up.sql
├── 003_create_guests.down.sql
├── 004_create_gifts.up.sql
├── 004_create_gifts.down.sql
├── 005_create_payments.up.sql
└── 005_create_payments.down.sql
```

Executadas automaticamente no boot via golang-migrate, ou manualmente:

```bash
make migrate-up
make migrate-down
make migrate-create name=add_column
```

## Índices

```sql
CREATE UNIQUE INDEX idx_weddings_slug ON weddings(slug);

CREATE INDEX idx_invitations_wedding_id ON invitations(wedding_id);
CREATE UNIQUE INDEX idx_invitations_wedding_code ON invitations(wedding_id, code);

CREATE INDEX idx_guests_invitation_id ON guests(invitation_id);
CREATE INDEX idx_guests_wedding_id ON guests(wedding_id);
CREATE INDEX idx_guests_wedding_status ON guests(wedding_id, status);
CREATE INDEX idx_guests_wedding_name ON guests(wedding_id, name COLLATE NOCASE);

CREATE INDEX idx_gifts_wedding_id ON gifts(wedding_id);
CREATE INDEX idx_gifts_category ON gifts(wedding_id, category);
CREATE INDEX idx_gifts_status ON gifts(wedding_id, status);

CREATE INDEX idx_payments_gift_id ON payments(gift_id);
CREATE INDEX idx_payments_wedding_id ON payments(wedding_id);
CREATE INDEX idx_payments_status ON payments(wedding_id, status);
CREATE INDEX idx_payments_provider_id ON payments(provider_id);
```

Índices compostos com `wedding_id` como prefixo garantem que queries scoped por tenant usem índices eficientemente.

## Arquivo do Banco

SQLite em arquivo único. Padrão: `./data/wedding.db`.

```
DATABASE_PATH=./data/wedding.db
```

### Nota sobre desnormalização de wedding_id

`guests` e `payments` possuem `wedding_id` mesmo sendo alcançável via `invitation_id → invitations.wedding_id` ou `gift_id → gifts.wedding_id`. Essa desnormalização intencional evita JOINs em queries frequentes que filtram por tenant, como listagens e dashboards.
