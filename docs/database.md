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
│ active           BOOLEAN     │                                  │
│ created_at       TIMESTAMPTZ │                                  │
│ updated_at       TIMESTAMPTZ │                                  │
└──────────────────────────────┘                                  │
         │                                                        │
         │  wedding_id FK                                         │
         ▼                                                        │
┌──────────────────────────┐       ┌───────────────────────────────────┐
│       invitations        │       │              guests               │
├──────────────────────────┤       ├───────────────────────────────────┤
│ id          TEXT PK      │◄──┐   │ id             TEXT PK            │
│ wedding_id  TEXT FK      │   │   │ invitation_id  TEXT FK            │
│ code        TEXT         │   └───│ wedding_id     TEXT FK            │
│ label       TEXT         │       │ name           TEXT               │
│ max_guests  INTEGER      │       │ phone          TEXT               │
│ notes       TEXT         │       │ email          TEXT               │
│ created_at  TIMESTAMPTZ  │       │ status         TEXT               │
│ updated_at  TIMESTAMPTZ  │       │ confirmed_at   TIMESTAMPTZ       │
└──────────────────────────┘       │ created_at     TIMESTAMPTZ       │
                                   │ updated_at     TIMESTAMPTZ       │
         │  wedding_id FK          └───────────────────────────────────┘
         ▼
┌───────────────────────────────────┐         ┌───────────────────────────────────┐
│              gifts                │         │            payments               │
├───────────────────────────────────┤         ├───────────────────────────────────┤
│ id            TEXT PK             │         │ id              TEXT PK           │
│ wedding_id    TEXT FK             │         │ wedding_id      TEXT FK           │
│ kind          TEXT                │         │ provider_id     TEXT              │
│ name, description, price, ...     │         │ amount (total)  DOUBLE PRECISION  │
│ image_url, category, status, ...    │         │ status, payment_method, ...      │
└───────────────────────────────────┘         └───────────────────────────────────┘
         ▲                                               ▲
         │ gift_id (linha: catálogo ou cash_template)    │ payment_id
         │         ┌─────────────────────────────────────┴──┐
         └─────────│            payment_items               │
                   ├────────────────────────────────────────┤
                   │ id, payment_id FK, gift_id FK, amount, │
                   │ custom_name, custom_description, ...   │
                   └────────────────────────────────────────┘
```

Cada cobrança tem uma ou mais linhas em `payment_items`. O total em `payments.amount` corresponde à soma das linhas. Presentes de catálogo referenciam `gifts` com `kind = catalog`; contribuição em dinheiro usa o único `gift` por casamento com `kind = cash_template` (id `cashttpl-<wedding_id>`).

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
| active | BOOLEAN | NOT NULL, DEFAULT TRUE | Ativo ou desativado |
| created_at | TIMESTAMPTZ | NOT NULL | Timestamp de criação |
| updated_at | TIMESTAMPTZ | NOT NULL | Timestamp da última atualização |

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
| created_at | TIMESTAMPTZ | NOT NULL | Timestamp de criação |
| updated_at | TIMESTAMPTZ | NOT NULL | Timestamp da última atualização |

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
| confirmed_at | TIMESTAMPTZ | | Timestamp da confirmação |
| created_at | TIMESTAMPTZ | NOT NULL | Timestamp de criação |
| updated_at | TIMESTAMPTZ | NOT NULL | Timestamp da última atualização |

### gifts

Linhas de **`kind = catalog`** são o catálogo exibido na lista de presentes. Existe **no máximo um** registro **`kind = cash_template` por wedding** (modelo de contribuição em dinheiro; não aparece na listagem de catálogo).

| Coluna | Tipo | Restrições | Descrição |
|--------|------|------------|-----------|
| id | TEXT | PK | UUID v4 ou `cashttpl-<wedding_id>` para o template de dinheiro |
| wedding_id | TEXT | FK → weddings(id), NOT NULL | Casamento ao qual pertence |
| kind | TEXT | NOT NULL | `catalog` ou `cash_template` |
| name | TEXT | NOT NULL | Nome do presente |
| description | TEXT | | Descrição detalhada |
| price | DOUBLE PRECISION | NOT NULL | Valor em reais (template de dinheiro pode ser 0) |
| image_url | TEXT | | URL da imagem |
| category | TEXT | NOT NULL | Categoria (ex: "Cozinha"); template pode usar `'cash'` |
| status | TEXT | NOT NULL, DEFAULT 'available' | `available` ou `purchased` |
| created_at | TIMESTAMPTZ | NOT NULL | Timestamp de criação |
| updated_at | TIMESTAMPTZ | NOT NULL | Timestamp da última atualização |

### payments

Cabeçalho da cobrança no gateway (valor total igual à soma de `payment_items`).

| Coluna | Tipo | Restrições | Descrição |
|--------|------|------------|-----------|
| id | TEXT | PK | UUID v4 |
| wedding_id | TEXT | FK → weddings(id), NOT NULL | Casamento (desnormalizado) |
| provider_id | TEXT | | ID da transação no provedor (InfinitePay slug ou Mercado Pago ID) |
| amount | DOUBLE PRECISION | NOT NULL | Valor total em reais |
| status | TEXT | NOT NULL, DEFAULT 'pending' | `pending`, `approved`, `rejected`, `expired` |
| payment_method | TEXT | NOT NULL | `pix` ou `credit_card` |
| payer_name | TEXT | NOT NULL | Nome de quem presenteia |
| payer_email | TEXT | | Email do pagador |
| message | TEXT | | Mensagem para os noivos |
| pix_qr_code | TEXT | | Código copia-e-cola do PIX |
| pix_expiration | TIMESTAMPTZ | | Quando o QR code expira |
| paid_at | TIMESTAMPTZ | | Timestamp do pagamento confirmado |
| created_at | TIMESTAMPTZ | NOT NULL | Timestamp de criação |
| updated_at | TIMESTAMPTZ | NOT NULL | Timestamp da última atualização |

### payment_items

Uma ou mais linhas por pagamento: presente de catálogo ou linha de contribuição em dinheiro (gift `cash_template`; texto opcional na linha).

| Coluna | Tipo | Restrições | Descrição |
|--------|------|------------|-----------|
| id | TEXT | PK | UUID da linha |
| payment_id | TEXT | FK → payments(id), NOT NULL | Pagamento pai |
| gift_id | TEXT | FK → gifts(id), NOT NULL | Gift do catálogo ou `cash_template` |
| amount | DOUBLE PRECISION | NOT NULL | Valor da linha |
| custom_name | TEXT | | Opcional para contribuição em dinheiro |
| custom_description | TEXT | | Opcional para contribuição em dinheiro |
| created_at | TIMESTAMPTZ | NOT NULL | Timestamp de criação |

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
├── … (001–008: weddings até ajustes de payments)
├── 009_payment_checkout.up.sql   # gifts.kind, payment_items, drop payments.gift_id
└── 009_payment_checkout.down.sql
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
CREATE INDEX idx_guests_wedding_name ON guests(wedding_id, LOWER(name));

CREATE INDEX idx_gifts_wedding_id ON gifts(wedding_id);
CREATE INDEX idx_gifts_category ON gifts(wedding_id, category);
CREATE INDEX idx_gifts_status ON gifts(wedding_id, status);
-- Um cash_template por wedding (parcial): ver migração 009

CREATE INDEX idx_payment_items_payment_id ON payment_items(payment_id);
CREATE INDEX idx_payment_items_gift_id ON payment_items(gift_id);

CREATE INDEX idx_payments_wedding_id ON payments(wedding_id);
CREATE INDEX idx_payments_status ON payments(wedding_id, status);
CREATE INDEX idx_payments_provider_id ON payments(provider_id);
```

Índices compostos com `wedding_id` como prefixo garantem que queries scoped por tenant usem índices eficientemente.

## Conexão

PostgreSQL via connection string. Compatível com qualquer provedor (Supabase, Neon, local).

```
DATABASE_URL=postgresql://user:pass@host:5432/wedding?sslmode=require
```

### Nota sobre desnormalização de wedding_id

`guests` e `payments` possuem `wedding_id` para filtro direto por tenant. Linhas detalham o que foi cobrado em `payment_items` (cada linha referencia um `gift_id`).
