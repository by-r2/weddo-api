# Configuração

Guia para preencher o arquivo `.env`. Copie o `.env.example` e ajuste os valores:

```bash
cp .env.example .env
```

## Servidor

| Variável | Descrição | Default | Obrigatório |
|----------|-----------|---------|-------------|
| `SERVER_PORT` | Porta HTTP | `8080` | Não |

## Banco de Dados

| Variável | Descrição | Default | Obrigatório |
|----------|-----------|---------|-------------|
| `DATABASE_PATH` | Caminho do arquivo SQLite | `./data/wedding.db` | Não |

O diretório é criado automaticamente. Em produção, use um caminho persistente fora do container.

## JWT (Autenticação Admin)

| Variável | Descrição | Default | Obrigatório |
|----------|-----------|---------|-------------|
| `JWT_SECRET` | Segredo para assinar tokens JWT | — | **Sim** |
| `JWT_EXPIRATION_HOURS` | Tempo de expiração do token em horas | `24` | Não |

### Como gerar o JWT_SECRET

Use um valor aleatório de pelo menos 32 bytes. **Nunca use valores previsíveis em produção.**

```bash
# Opção 1 — openssl (recomendado)
openssl rand -base64 32

# Opção 2 — /dev/urandom
head -c 32 /dev/urandom | base64

# Opção 3 — python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

Copie o valor gerado e cole no `.env`:

```
JWT_SECRET=K7xP3nQ9wR2vL5mJ8dF1hG6tY0uI4oA3sE7bN2cX9z=
```

## Wedding Seed (Primeiro Tenant)

Variáveis para criar automaticamente o primeiro casamento no boot. Se `SEED_ADMIN_EMAIL` e `SEED_ADMIN_PASSWORD` estiverem vazios, o seed é ignorado.

| Variável | Descrição | Exemplo | Obrigatório |
|----------|-----------|---------|-------------|
| `SEED_WEDDING_SLUG` | Slug para uso no frontend | `manu-rafa` | Não* |
| `SEED_WEDDING_TITLE` | Título do casamento (formato: `Casamento Nome & Nome`) | `Casamento Manoela & Rafael` | Não* |
| `SEED_WEDDING_DATE` | Data do casamento (`YYYY-MM-DD`) | `2026-07-07` | Não* |
| `SEED_ADMIN_EMAIL` | Email do administrador | `admin@manurafa.com.br` | Não* |
| `SEED_ADMIN_PASSWORD` | Senha do administrador (texto plano → armazenada como bcrypt) | — | Não* |

*Obrigatórios apenas se quiser criar o tenant inicial automaticamente.

### Como escolher uma senha segura para SEED_ADMIN_PASSWORD

```bash
# Gerar senha aleatória de 20 caracteres
openssl rand -base64 20

# Ou uma passphrase memorável
python3 -c "import secrets; print('-'.join(secrets.choice('abcdefghijklmnopqrstuvwxyz') + secrets.token_hex(2) for _ in range(4)))"
```

> A senha é hasheada com **bcrypt** antes de ser armazenada. O valor em texto plano existe apenas no `.env` (que é gitignored).

## Pagamentos (Lista de Presentes)

A API suporta dois provedores de pagamento. Escolha um via `PAYMENT_PROVIDER`:

| Variável | Descrição | Default | Obrigatório |
|----------|-----------|---------|-------------|
| `PAYMENT_PROVIDER` | Provedor ativo: `infinitepay` ou `mercadopago` | (vazio = desabilitado) | Sim* |

*Se vazio, endpoints de pagamento retornam `503 Service Unavailable`.

### InfinitePay (recomendado — taxas menores)

| Taxa | Valor |
|------|-------|
| PIX | **0%** |
| Crédito à vista | ~2,69% |
| Crédito 12x | ~8,99% |

| Variável | Descrição | Obrigatório |
|----------|-----------|-------------|
| `IP_HANDLE` | Sua InfiniteTag (nome de usuário no app, sem o `$`) | Sim |
| `IP_REDIRECT_URL` | URL para onde o comprador volta após pagar | Não |
| `IP_WEBHOOK_URL` | URL que receberá notificações de pagamento | Não |

**Fluxo**: checkout por redirect — o comprador é enviado para a tela da InfinitePay, paga (PIX ou cartão), e volta ao site.

#### Como obter a InfiniteTag

1. Baixe o app [InfinitePay](https://www.infinitepay.io/) e crie uma conta
2. Sua **InfiniteTag** fica em **Configurações > Perfil** (ex: `manu-rafa`)
3. Use sem o símbolo `$`: `IP_HANDLE=manu-rafa`

#### Webhook

Configure `IP_WEBHOOK_URL` com a URL pública da sua API:

```
IP_WEBHOOK_URL=https://api.manurafa.com.br/api/v1/payments/webhook
```

A InfinitePay envia um POST quando o pagamento é aprovado. A API atualiza automaticamente o status do presente.

### Mercado Pago (alternativa — checkout transparente)

| Taxa | Valor |
|------|-------|
| PIX | ~0,99% |
| Crédito à vista | ~4,98% |
| Crédito 12x | ~14-17% |

| Variável | Descrição | Obrigatório |
|----------|-----------|-------------|
| `MP_ACCESS_TOKEN` | Token de acesso à API do Mercado Pago | Sim |
| `MP_WEBHOOK_SECRET` | Segredo para validar assinatura dos webhooks | Sim |
| `MP_NOTIFICATION_URL` | URL pública que receberá notificações de pagamento | Sim |
| `MP_PIX_EXPIRATION_MINUTES` | Tempo de expiração do QR Code PIX | Não (default: 30) |

**Fluxo**: checkout transparente — o pagador não sai do site. QR Code PIX e dados do cartão são processados inline.

#### Como obter as credenciais do Mercado Pago

1. Acesse [mercadopago.com.br/developers](https://www.mercadopago.com.br/developers)
2. Crie uma conta ou faça login
3. Vá em **Suas integrações** → **Criar aplicação**
4. Preencha o nome (ex: `Mr Wedding`) e selecione **Checkout API** como produto
5. Após criar, acesse **Credenciais de produção**:
   - **Access Token** → copie para `MP_ACCESS_TOKEN`
6. Para o webhook secret, vá em **Webhooks** → configure a URL de notificação
   - A URL deve ser pública (ex: `https://api.manurafa.com.br/api/v1/payments/webhook`)
   - Copie o segredo gerado para `MP_WEBHOOK_SECRET`

#### Ambiente de teste (sandbox)

Para desenvolvimento, use as **credenciais de teste** (disponíveis na mesma página de credenciais). Elas operam num ambiente sandbox onde pagamentos não são reais.

```
MP_ACCESS_TOKEN=TEST-xxxxxxxxxxxx
```

## CORS

| Variável | Descrição | Default | Obrigatório |
|----------|-----------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | Origens permitidas, separadas por vírgula | `*` | Não |

```bash
# Desenvolvimento
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5500

# Produção
CORS_ALLOWED_ORIGINS=https://manurafa.com.br,https://www.manurafa.com.br
```

## Logging

| Variável | Descrição | Default | Obrigatório |
|----------|-----------|---------|-------------|
| `LOG_LEVEL` | Nível de log: `debug`, `info`, `warn`, `error` | `info` | Não |
| `LOG_FORMAT` | Formato do log: `text` (legível) ou `json` (estruturado) | `text` | Não |

Use `debug` em desenvolvimento para ver detalhes de requests. Em produção, `info` ou `warn`.

Em produção use `LOG_FORMAT=json` para logs estruturados compatíveis com sistemas de observabilidade (Datadog, Grafana Loki, etc.).

## Postman CLI

| Variável | Descrição | Default | Obrigatório |
|----------|-----------|---------|-------------|
| `POSTMAN_API_KEY` | API Key para sincronizar collection/environment com o Postman Cloud | — | Não* |

*Obrigatório apenas para `make postman-push`. Não é usado pela aplicação Go.

### Como gerar a API Key

1. Acesse [go.postman.co/settings/me/api-keys](https://go.postman.co/settings/me/api-keys)
2. Clique em **Generate API Key**
3. Dê um nome (ex: `wedding-api-ci`)
4. Copie a chave (formato `PMAK-...`) e cole no `.env`:

```
POSTMAN_API_KEY=PMAK-xxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx
```

O CI do GitHub Actions usa essa mesma chave via secret do repositório (não lê do `.env`).

## Exemplos completos (.env de produção)

### Com InfinitePay

```bash
SERVER_PORT=8080
DATABASE_PATH=/var/data/wedding.db

JWT_SECRET=<valor gerado com openssl rand -base64 32>
JWT_EXPIRATION_HOURS=12

SEED_WEDDING_SLUG=manu-rafa
SEED_WEDDING_TITLE=Casamento Manoela & Rafael
SEED_WEDDING_DATE=2026-07-07
SEED_ADMIN_EMAIL=admin@manurafa.com.br
SEED_ADMIN_PASSWORD=<senha segura gerada>

PAYMENT_PROVIDER=infinitepay
IP_HANDLE=manu-rafa
IP_REDIRECT_URL=https://manurafa.com.br/obrigado
IP_WEBHOOK_URL=https://api.manurafa.com.br/api/v1/payments/webhook

CORS_ALLOWED_ORIGINS=https://manurafa.com.br,https://www.manurafa.com.br

LOG_LEVEL=info
LOG_FORMAT=json
```

### Com Mercado Pago

```bash
SERVER_PORT=8080
DATABASE_PATH=/var/data/wedding.db

JWT_SECRET=<valor gerado com openssl rand -base64 32>
JWT_EXPIRATION_HOURS=12

SEED_WEDDING_SLUG=manu-rafa
SEED_WEDDING_TITLE=Casamento Manoela & Rafael
SEED_WEDDING_DATE=2026-07-07
SEED_ADMIN_EMAIL=admin@manurafa.com.br
SEED_ADMIN_PASSWORD=<senha segura gerada>

PAYMENT_PROVIDER=mercadopago
MP_ACCESS_TOKEN=APP_USR-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
MP_WEBHOOK_SECRET=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
MP_NOTIFICATION_URL=https://api.manurafa.com.br/api/v1/payments/webhook
MP_PIX_EXPIRATION_MINUTES=30

CORS_ALLOWED_ORIGINS=https://manurafa.com.br,https://www.manurafa.com.br

LOG_LEVEL=info
LOG_FORMAT=json
```
