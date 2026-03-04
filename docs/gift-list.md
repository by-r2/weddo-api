# Lista de Presentes

## Como Funciona

A lista de presentes de casamento funciona como um catálogo de **presentes virtuais**. Os itens representam presentes reais (eletrodomésticos, lua de mel, etc.), mas o convidado paga em dinheiro e o valor vai direto para a conta dos noivos. O casal usa o dinheiro como quiser.

### Fluxo do Convidado

```
Acessa a lista → Escolhe presente → Clica "Presentear"
  → Informa nome e mensagem (opcional)
  → Escolhe forma de pagamento (PIX ou cartão)
    → PIX: recebe QR code, paga no app do banco
    → Cartão: preenche dados no checkout
  → API recebe webhook do Mercado Pago confirmando pagamento
  → Presente marcado como "comprado"
  → Convidado vê tela de agradecimento
```

### Fluxo do Admin (noivos)

```
Cadastra presentes (nome, descrição, preço, imagem, categoria)
  → Publica a lista
  → Acompanha presentes comprados e valores recebidos
  → Dinheiro cai na conta Mercado Pago → transfere para banco
```

## Gateway de Pagamento: Mercado Pago

### Por que Mercado Pago

- SDK oficial para Go ([mercadopago/sdk-go](https://github.com/mercadopago/sdk-go))
- PIX com taxa de **~0.5%** (a menor do mercado)
- Cartão de crédito com taxa de **~4-5%**
- Amplamente usado no Brasil
- Webhooks confiáveis para confirmação de pagamento
- Checkout Transparente (todo o fluxo no nosso site, sem redirecionar)

### Alternativa Considerada

- **Stripe**: suporta PIX no Brasil, mas taxas maiores e menos comum entre brasileiros.

### Configuração Necessária

1. Criar conta no [Mercado Pago](https://www.mercadopago.com.br/)
2. Gerar credenciais de teste (sandbox) e produção
3. Cadastrar chave PIX na conta
4. Configurar URL de webhook no painel do MP

### Variáveis de Ambiente

```
MP_ACCESS_TOKEN=APP_USR-xxxx          # Token de acesso (sandbox ou produção)
MP_WEBHOOK_SECRET=xxxx                # Secret para validar webhooks
MP_NOTIFICATION_URL=https://api.dominio.com/api/v1/payments/webhook
```

## Onde Cada Coisa Acontece

| Responsabilidade | Camada |
|------------------|--------|
| Exibir catálogo de presentes | Frontend |
| Exibir formulário com nome/mensagem | Frontend |
| Criar pagamento (gerar PIX QR code) | **Backend** (via API do Mercado Pago) |
| Exibir QR code PIX para o convidado | Frontend (recebe do backend) |
| Processar checkout com cartão | **Frontend** (SDK JS do Mercado Pago) + **Backend** (criar pagamento) |
| Confirmar pagamento via webhook | **Backend** (recebe notificação do Mercado Pago) |
| Marcar presente como comprado | **Backend** |
| CRUD de presentes | **Backend** (admin) |
| Relatório financeiro | **Backend** (admin) |

## Fluxo Técnico — PIX

```
Frontend                     Backend                      Mercado Pago
   │                            │                              │
   │  POST /gifts/:id/purchase  │                              │
   │  { name, message }         │                              │
   │───────────────────────────>│                              │
   │                            │  POST /v1/payments           │
   │                            │  { amount, payer, pix... }   │
   │                            │─────────────────────────────>│
   │                            │                              │
   │                            │  { id, qr_code, qr_code_    │
   │                            │    base64, ticket_url }      │
   │                            │<─────────────────────────────│
   │                            │                              │
   │  { payment_id, qr_code,   │                              │
   │    qr_code_base64,         │                              │
   │    expiration }             │                              │
   │<───────────────────────────│                              │
   │                            │                              │
   │  [Convidado paga no app]   │                              │
   │                            │                              │
   │                            │  POST /webhook (notification)│
   │                            │  { action: payment.updated } │
   │                            │<─────────────────────────────│
   │                            │                              │
   │                            │  GET /v1/payments/:id        │
   │                            │─────────────────────────────>│
   │                            │  { status: approved }        │
   │                            │<─────────────────────────────│
   │                            │                              │
   │                            │  [Marca gift como purchased] │
   │                            │  [Registra payment]          │
```

## Fluxo Técnico — Cartão de Crédito

O cartão de crédito usa o Checkout Transparente do Mercado Pago. O frontend coleta os dados do cartão via SDK JS do MP (que tokeniza no lado do cliente, sem dados sensíveis passarem pelo nosso backend).

```
Frontend (SDK JS do MP)      Backend                      Mercado Pago
   │                            │                              │
   │  [Tokeniza cartão via      │                              │
   │   MercadoPago.js]          │                              │
   │────────────────────────────────────────────────────────-->│
   │  { card_token }            │                              │
   │<──────────────────────────────────────────────────────────│
   │                            │                              │
   │  POST /gifts/:id/purchase  │                              │
   │  { name, message,          │                              │
   │    card_token,              │                              │
   │    payment_method_id,       │                              │
   │    installments }           │                              │
   │───────────────────────────>│                              │
   │                            │  POST /v1/payments           │
   │                            │  { token, amount, ... }      │
   │                            │─────────────────────────────>│
   │                            │  { status: approved }        │
   │                            │<─────────────────────────────│
   │                            │                              │
   │  { status: approved }      │                              │
   │<───────────────────────────│                              │
```

## Segurança

- **Cartão**: nunca toca nosso backend. O SDK JS do Mercado Pago tokeniza no cliente.
- **Webhook**: validar assinatura do Mercado Pago antes de processar.
- **Idempotência**: usar `idempotency_key` ao criar pagamentos para evitar duplicatas.
- **Expiração PIX**: QR code expira em 30 minutos. Se expirar, o presente volta a ficar disponível.
