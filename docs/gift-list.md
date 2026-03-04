# Lista de Presentes

## Como Funciona

A lista de presentes de casamento funciona como um catálogo de **presentes virtuais**. Os itens representam presentes reais (eletrodomésticos, lua de mel, etc.), mas o convidado paga em dinheiro e o valor vai direto para a conta dos noivos. O casal usa o dinheiro como quiser.

### Fluxo do Convidado

```
Acessa a lista → Escolhe presente → Clica "Presentear"
  → Informa nome e mensagem (opcional)
  → É direcionado ao checkout do provedor (InfinitePay ou Mercado Pago)
  → Paga (PIX ou cartão)
  → Webhook confirma pagamento → presente marcado como "comprado"
  → Convidado vê tela de agradecimento
```

### Fluxo do Admin (noivos)

```
Cadastra presentes (nome, descrição, preço, imagem, categoria)
  → Publica a lista
  → Acompanha presentes comprados e valores recebidos
  → Dinheiro cai na conta do provedor → transfere para banco
```

## Provedores de Pagamento

A API suporta dois provedores via `PAYMENT_PROVIDER`. A interface `PaymentGateway` abstrai as diferenças.

### Comparação

| Aspecto | InfinitePay | Mercado Pago |
|---------|-------------|--------------|
| **PIX** | **0%** | ~0,99% |
| **Crédito à vista** | ~2,69% | ~4,98% |
| **Crédito 12x** | ~8,99% | ~14-17% |
| **Fluxo** | Redirect (checkout externo) | Transparente (inline no site) |
| **Webhook** | Sim | Sim |
| **SDK Go** | Não (API REST simples) | Sim (v1.8.0) |
| **Ideal para** | Economia | UX premium |

### InfinitePay (recomendado)

Taxas menores, especialmente PIX grátis. O comprador é redirecionado para o checkout da InfinitePay, paga, e volta ao site.

**Variáveis**: `IP_HANDLE`, `IP_REDIRECT_URL`, `IP_WEBHOOK_URL`

### Mercado Pago

Checkout transparente — tudo acontece sem sair do site. QR Code PIX inline e tokenização de cartão via SDK JS.

**Variáveis**: `MP_ACCESS_TOKEN`, `MP_WEBHOOK_SECRET`, `MP_NOTIFICATION_URL`, `MP_PIX_EXPIRATION_MINUTES`

## Onde Cada Coisa Acontece

| Responsabilidade | Camada |
|------------------|--------|
| Exibir catálogo de presentes | Frontend |
| Exibir formulário com nome/mensagem | Frontend |
| Criar pagamento (link checkout ou QR code) | **Backend** (via gateway) |
| Redirecionar para checkout (InfinitePay) | Frontend (recebe `checkout_url`) |
| Exibir QR code PIX (Mercado Pago) | Frontend (recebe `qr_code`) |
| Confirmar pagamento via webhook | **Backend** (recebe notificação do provedor) |
| Marcar presente como comprado | **Backend** |
| CRUD de presentes | **Backend** (admin) |
| Relatório financeiro | **Backend** (admin) |

## Fluxo Técnico — InfinitePay

```
Frontend                     Backend                      InfinitePay
   │                            │                              │
   │  POST /gifts/:id/purchase  │                              │
   │  { name, email, message,   │                              │
   │    payment_method: "pix" }  │                              │
   │───────────────────────────>│                              │
   │                            │  POST /checkout/links        │
   │                            │  { handle, items, webhook }  │
   │                            │─────────────────────────────>│
   │                            │  { url, slug }               │
   │                            │<─────────────────────────────│
   │                            │                              │
   │  { payment_id,             │                              │
   │    checkout_url: "..." }   │                              │
   │<───────────────────────────│                              │
   │                            │                              │
   │  [Redireciona → paga]      │                              │
   │                            │                              │
   │                            │  POST /webhook               │
   │                            │  { order_nsu, paid_amount }  │
   │                            │<─────────────────────────────│
   │                            │                              │
   │                            │  [Marca gift como purchased] │
```

## Fluxo Técnico — Mercado Pago (PIX)

```
Frontend                     Backend                      Mercado Pago
   │                            │                              │
   │  POST /gifts/:id/purchase  │                              │
   │  { name, email, message,   │                              │
   │    payment_method: "pix" }  │                              │
   │───────────────────────────>│                              │
   │                            │  POST /v1/payments           │
   │                            │  { amount, payer, pix... }   │
   │                            │─────────────────────────────>│
   │                            │  { id, qr_code, ... }        │
   │                            │<─────────────────────────────│
   │                            │                              │
   │  { payment_id, qr_code,   │                              │
   │    qr_code_base64,         │                              │
   │    expiration }             │                              │
   │<───────────────────────────│                              │
   │                            │                              │
   │  [Convidado paga no app]   │                              │
   │                            │                              │
   │                            │  POST /webhook               │
   │                            │  { action: payment.updated } │
   │                            │<─────────────────────────────│
   │                            │                              │
   │                            │  [Marca gift como purchased] │
```

## Fluxo Técnico — Mercado Pago (Cartão)

O cartão de crédito usa o Checkout Transparente. O frontend coleta os dados do cartão via SDK JS do MP (tokeniza no cliente, sem dados sensíveis no backend).

```
Frontend (SDK JS do MP)      Backend                      Mercado Pago
   │                            │                              │
   │  [Tokeniza cartão]        │                              │
   │────────────────────────────────────────────────────────-->│
   │  { card_token }            │                              │
   │<──────────────────────────────────────────────────────────│
   │                            │                              │
   │  POST /gifts/:id/purchase  │                              │
   │  { card_token, ... }       │                              │
   │───────────────────────────>│                              │
   │                            │  POST /v1/payments           │
   │                            │─────────────────────────────>│
   │                            │  { status: approved }        │
   │                            │<─────────────────────────────│
   │                            │                              │
   │  { status: approved }      │                              │
   │<───────────────────────────│                              │
```

## Comportamento do Frontend

O frontend deve verificar a resposta do `POST /purchase`:

- Se `checkout_url` está presente → **redirecionar** o usuário para essa URL (InfinitePay)
- Se `qr_code` está presente → **exibir** o QR code inline (Mercado Pago PIX)
- Se `status: approved` → **exibir** tela de sucesso imediato (Mercado Pago cartão)

## Segurança

- **Cartão (Mercado Pago)**: nunca toca nosso backend. O SDK JS do MP tokeniza no cliente.
- **Webhook**: validar origem antes de processar (IP ou assinatura).
- **Expiração PIX**: QR code expira em 30 minutos (Mercado Pago). Se expirar, o presente volta a ficar disponível.
- **Idempotência**: `external_reference` / `order_nsu` previnem duplicatas.
