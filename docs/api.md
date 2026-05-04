# API — Endpoints e Contratos

Base URL: `http://localhost:8080/api/v1`

Todas as respostas seguem o formato JSON. Erros retornam o campo `error`.

## Multi-tenancy

A API é multi-tenant. Cada casamento é um tenant isolado.

- **Endpoints públicos**: o tenant é identificado pelo `{weddingId}` (UUID) na URL
- **Endpoints admin**: o tenant é extraído do JWT (campo `wedding_id` nos claims)
- **Webhook**: o tenant é resolvido via dados do pagamento no banco

Se o `weddingId` não existir ou o wedding estiver inativo, retorna `404`. O frontend é responsável por conhecer o UUID do seu wedding (configurado como variável de ambiente ou build-time).

## Autenticação

Endpoints admin (`/api/v1/admin/*`) exigem header `Authorization: Bearer <token>`.
O token é obtido via `POST /api/v1/admin/auth` e contém `wedding_id` nos claims.

---

## Endpoints Públicos

Prefixo: `/api/v1/w/{weddingId}`

### Confirmar Presença (RSVP)

```
POST /api/v1/w/{weddingId}/rsvp
```

**Request:**

```json
{
  "code": "SILVA-001",
  "name": "João Silva",
  "will_attend": true
}
```

- `will_attend` (opcional): se omitido ou `true`, confirma presença; se `false`, registra que o convidado não comparecerá (`status` `declined`).
- O `code` é o mesmo campo do convite em `invitations` (único por casamento), usado também no `GET` abaixo e no painel.

**Response 200 — confirmação registrada:**

```json
{
  "guest": {
    "id": "uuid",
    "name": "João Silva",
    "status": "confirmed",
    "confirmed_at": "2026-03-04T10:30:00Z"
  },
  "invitation": {
    "label": "Família Silva"
  },
  "message": "Presença confirmada com sucesso!"
}
```

**Response 404 — convite ou convidado não encontrado:**

Mensagens possíveis: `Convite não encontrado.` ou `Convidado não encontrado neste convite. Verifique o nome.`

**Response 409 — já confirmado:**

```json
{
  "guest": {
    "id": "uuid",
    "name": "João Silva",
    "status": "confirmed",
    "confirmed_at": "2026-03-01T14:00:00Z"
  },
  "invitation": {
    "label": "Família Silva"
  },
  "message": "Presença já estava confirmada."
}
```

**Response 409 — convidado recusado anteriormente:**

```json
{
  "error": "Este convidado recusou o convite e não pode confirmar novamente."
}
```

**Response 200 — recusa registrada (`will_attend: false`):**

O corpo inclui `guest.status` `"declined"` e `confirmed_at` nulo. Mensagens exemplares: confirmação da recusa ou, se já estava recusado, agradecimento idempotente (sem novo `UPDATE`).

### Consultar convite (lista do grupo)

```
GET /api/v1/w/{weddingId}/rsvp/invitation?code=SILVA-001
```

**Response 200:**

```json
{
  "invitation": {
    "label": "Família Silva",
    "max_guests": 4
  },
  "guests": [
    { "id": "uuid", "name": "João Silva", "status": "confirmed" },
    { "id": "uuid", "name": "Maria Silva", "status": "pending" },
    { "id": "uuid", "name": "Pedro Silva", "status": "pending" }
  ]
}
```

### Categorias de presentes (para select / filtro)

Categorias já usadas em `category` nos gifts de **catálogo** do casamento (`kind = catalog`) com total por categoria. Ordenação alfabética case-insensitive. Se ainda não existir nenhum presente, `categories` vem vazio.

```
GET /api/v1/w/{weddingId}/gift-categories
GET /api/v1/admin/gift-categories
```

O endpoint **admin** usa o `wedding_id` do JWT (sem path extra).

**Response 200:**

```json
{
  "categories": [
    { "name": "Cozinha", "count": 12 },
    { "name": "Lua de mel", "count": 8 },
    { "name": "Quarto", "count": 5 }
  ]
}
```

### Listar Presentes

```
GET /api/v1/w/{weddingId}/gifts?page=1&per_page=20&category=Cozinha&category=Quarto&search=panela&min_price=0&max_price=500&sort_by=recommended&sort_dir=asc
```

**Query params opcionais**

| Parâmetro | Descrição |
|-----------|-----------|
| `page`, `per_page` | Paginação (`per_page` máximo 100 no servidor). |
| `category` | Repita o parâmetro para várias categorias: `?category=A&category=B`. Match exato ao valor salvo no presente. |
| `search` | Busca por nome (`ILIKE`). |
| `min_price`, `max_price` | Faixa de preço (números ≥ 0; `min_price` não pode ser maior que `max_price`). |
| `sort_by` | `recommended` (padrão: `category` ASC, `name` ASC), `price` ou `name`. |
| `sort_dir` | `asc` ou `desc` (padrão `asc`). Ignorado quando `sort_by=recommended`. |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Jogo de Panelas",
      "description": "Jogo com 5 peças antiaderente",
      "price": 350.00,
      "image_url": "https://...",
      "category": "Cozinha",
      "status": "available",
      "created_at": "2026-03-01T10:00:00Z",
      "updated_at": "2026-03-01T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 30, "total_pages": 2 }
}
```

A listagem pública mostra apenas presentes do **catálogo** com status `available`. O modelo interno **Contribuição em dinheiro** (`kind: cash_template`, id `cashttpl-<uuid do casamento>`) **não** entra nesta lista — aparece apenas no fluxo de checkout e nos pagamentos. Filtros opcionais: ver tabela acima (`category` repetido, `search`, `min_price`, `max_price`, `sort_by`, `sort_dir`, paginação).

### Detalhar Presente

```
GET /api/v1/w/{weddingId}/gifts/{id}
```

### Checkout — iniciar pagamento (lista de presentes no corpo)

Carrinho apenas no cliente: o front envia **todos os itens** em um único pedido (`items[]`). Presentes do **catálogo** entram apenas com `gift_id` (valor vem da lista da loja). A **contribuição em dinheiro** usa o `gift_id` do modelo “Contribuição em dinheiro” (criado pelo sistema por casamento) e exige `amount`; opcionalmente `custom_name` e `custom_description`.

```
POST /api/v1/w/{weddingId}/checkout
```

**Request (exemplo: um presente de catálogo + contribuição em dinheiro):**

```json
{
  "items": [
    { "gift_id": "uuid-do-presente-catalogo" },
    {
      "gift_id": "cashttpl-<wedding_uuid>",
      "amount": 150.50,
      "custom_name": "Para a lua de mel",
      "custom_description": "Com carinho ❤️"
    }
  ],
  "payer_name": "Tia Maria",
  "payer_email": "maria@email.com",
  "message": "Felicidades ao casal!",
  "payment_method": "pix",
  "redirect_url": "https://manurafa.com.br/obrigado"
}
```

Para **apenas um presente da lista**, envie um único item em `items`. Não envie `amount` nem texto personalizado em itens do catálogo.

> `redirect_url` é opcional — sobrescreve o `IP_REDIRECT_URL` global (apenas InfinitePay).
> `card_token`, `payment_method_id` e `installments` são obrigatórios apenas para `credit_card` com Mercado Pago.

**Response 201 — InfinitePay (checkout redirect):**

```json
{
  "payment_id": "uuid",
  "provider_id": "slug-abc123",
  "status": "pending",
  "checkout_url": "https://checkout.infinitepay.io/abc123"
}
```

> Frontend deve redirecionar o usuário para `checkout_url`.

**Response 201 — Mercado Pago PIX (QR code inline):**

```json
{
  "payment_id": "uuid",
  "provider_id": "mp-123456",
  "status": "pending",
  "qr_code": "00020126...",
  "qr_code_base64": "data:image/png;base64,...",
  "expires_at": "2026-03-04T11:00:00Z"
}
```

**Response 201 — Mercado Pago cartão (aprovação imediata):**

```json
{
  "payment_id": "uuid",
  "provider_id": "mp-123456",
  "status": "approved"
}
```

**Lógica do frontend:**
- Se `checkout_url` presente → redirecionar
- Se `qr_code` presente → exibir QR code
- Se `status: approved` → tela de sucesso

### Consultar Status do Pagamento

Polling enquanto aguarda pagamento.

```
GET /api/v1/w/{weddingId}/payments/{id}/status
```

**Response 200:**

```json
{
  "payment_id": "uuid",
  "status": "approved",
  "lines": [
    { "gift_id": "uuid", "kind": "catalog", "amount": 350, "label": "Jogo de panelas" },
    { "gift_id": "cashttpl-…", "kind": "cash_template", "amount": 150.5, "label": "Para a lua de mel" }
  ]
}
```

### Webhook de Pagamento

Recebe notificações do provedor ativo. Não é chamado pelo frontend. O tenant é resolvido internamente via pagamento persistido (`payment_items` + `payments`).

```
POST /api/v1/payments/webhook
```

O formato do payload depende do provedor configurado em `PAYMENT_PROVIDER`. A API detecta e processa automaticamente.

---

## Endpoints Admin

Prefixo: `/api/v1/admin`

O `wedding_id` vem do JWT — não precisa de slug na URL.

### Login

```
POST /api/v1/admin/auth
```

**Request:**

```json
{
  "email": "manu.rafa@email.com",
  "password": "senha"
}
```

**Response 200:**

```json
{
  "token": "eyJhbGciOi...",
  "wedding": {
    "id": "uuid",
    "slug": "manoela-rafael",
    "title": "Casamento Manoela & Rafael"
  }
}
```

### Dashboard

```
GET /api/v1/admin/dashboard
```

**Response 200:**

```json
{
  "rsvp": {
    "total_invitations": 80,
    "total_guests": 200,
    "confirmed": 120,
    "pending": 75,
    "declined": 5,
    "confirmation_rate": 61.5
  },
  "gifts": {
    "total_gifts": 50,
    "purchased": 18,
    "available": 32,
    "total_revenue": 6500.00,
    "total_payments": 18
  }
}
```

> O campo `gifts` só aparece quando há presentes cadastrados.

### Convites (Invitations)

```
GET    /api/v1/admin/invitations          # listar (?page=1&per_page=20&search=silva)
POST   /api/v1/admin/invitations          # criar
GET    /api/v1/admin/invitations/{id}     # detalhar (inclui guests)
PUT    /api/v1/admin/invitations/{id}     # atualizar
DELETE /api/v1/admin/invitations/{id}     # remover (cascade guests)
```

**Criar convite com convidados (schema completo):**

```json
{
  "label": "Família Silva",
  "max_guests": 4,
  "notes": "Mesa próxima à família",
  "guests": [
    {
      "name": "João Silva",
      "phone": "11999998888",
      "email": "joao@email.com",
      "status": "pending"
    },
    {
      "name": "Maria Silva",
      "phone": "11999997777",
      "email": "maria@email.com",
      "status": "confirmed"
    }
  ]
}
```

> O `code` é gerado automaticamente pela API (alfanumérico curto com letras maiúsculas e números em ordem aleatória).
> `status` é opcional em cada convidado (`pending`, `confirmed` ou `declined`). Se omitido, usa `pending`.

**Detalhar convite (response):**

```json
{
  "id": "uuid",
  "code": "SILVA-001",
  "label": "Família Silva",
  "max_guests": 4,
  "guests": [
    { "id": "uuid", "name": "João Silva", "status": "confirmed", "confirmed_at": "..." },
    { "id": "uuid", "name": "Maria Silva", "status": "pending", "confirmed_at": null },
    { "id": "uuid", "name": "Pedro Silva", "status": "pending", "confirmed_at": null }
  ],
  "created_at": "2026-03-01T10:00:00Z",
  "updated_at": "2026-03-01T10:00:00Z"
}
```

### Convidados (Guests)

```
GET    /api/v1/admin/guests               # listar (?page=1&per_page=20&status=confirmed&search=joão)
GET    /api/v1/admin/guests/{id}          # detalhar
PUT    /api/v1/admin/guests/{id}          # atualizar
DELETE /api/v1/admin/guests/{id}          # remover
POST   /api/v1/admin/invitations/{id}/guests  # adicionar a convite existente
```

**Adicionar convidado a convite existente:**

```json
{
  "name": "Ana Silva",
  "phone": "11988887777",
  "email": "ana@email.com",
  "status": "pending"
}
```

> `status` também é opcional aqui e segue o mesmo comportamento (default `pending`).

**Regras de transição de status (`PUT /api/v1/admin/guests/{id}`):**

- `pending -> confirmed` ✅
- `pending -> declined` ✅
- `confirmed -> declined` ✅
- `confirmed -> pending` ❌
- `declined -> pending` ❌
- `declined -> confirmed` ❌

### Presentes (Gifts)

Lista e CRUD apenas para **`kind = catalog`** (o modelo de contribuição em dinheiro não aparece nem pode ser alterado por aqui).

```
GET    /api/v1/admin/gifts                # listar (?page=&per_page=&category=&status=&search=&min_price=&max_price=&sort_by=&sort_dir=)
POST   /api/v1/admin/gifts                # criar
GET    /api/v1/admin/gifts/{id}           # detalhar
PUT    /api/v1/admin/gifts/{id}           # atualizar
DELETE /api/v1/admin/gifts/{id}           # remover
```

### Pagamentos

```
GET /api/v1/admin/payments                # listar (?page=1&per_page=20&status=approved&gift_id=uuid)
GET /api/v1/admin/payments/{id}           # detalhar
```

### Google Sheets (sincronização manual)

```
POST /api/v1/admin/sheets/connect/start   # inicia OAuth e retorna auth_url
POST /api/v1/admin/sheets/push            # exporta banco -> planilha
POST /api/v1/admin/sheets/pull            # importa planilha -> banco
GET  /api/v1/sheets/connect/callback      # callback OAuth do Google
```

#### Iniciar conexão OAuth

`POST /api/v1/admin/sheets/connect/start`

**Response 200:**

```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/auth?..."
}
```

> O frontend/admin deve redirecionar o usuário para `auth_url`.

#### Push (banco para planilha)

`POST /api/v1/admin/sheets/push`

**Response 200:**

```json
{
  "invitations": 42,
  "guests": 160,
  "gifts": 28,
  "payments": 19
}
```

> Cada tenant usa sua própria planilha (conectada via OAuth).

#### Pull (planilha para banco)

`POST /api/v1/admin/sheets/pull`

**Response 200:**

```json
{
  "invitations_updated": 3,
  "invitations_created": 1,
  "guests_updated": 8,
  "guests_created": 2,
  "skipped": 4
}
```

> O pull atualiza/cria apenas dados de **convites** e **convidados**. Abas de presentes/pagamentos são somente exportação.

#### Callback OAuth

`GET /api/v1/sheets/connect/callback?code=...&state=...`

**Response 200:**

```json
{
  "wedding_id": "uuid",
  "spreadsheet_id": "1AbCdEf...",
  "spreadsheet_url": "https://docs.google.com/spreadsheets/d/1AbCdEf.../edit"
}
```

---

## Padrão de Resposta de Erro

```json
{
  "error": "mensagem descritiva do erro"
}
```

| Status | Uso |
|--------|-----|
| 400 | Validação falhou ou request malformado |
| 401 | Token ausente ou inválido |
| 404 | Recurso não encontrado ou wedding_id inválido |
| 409 | Conflito (ex: presença já confirmada, presente indisponível) |
| 500 | Erro interno |
| 503 | Serviço indisponível (ex: `PAYMENT_PROVIDER` ou Google Sheets não configurado) |

---

## Paginação

Endpoints de listagem suportam paginação:

```
GET /api/v1/admin/guests?page=1&per_page=20
```

```json
{
  "data": [...],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 200,
    "total_pages": 10
  }
}
```
