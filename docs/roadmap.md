# Roadmap

## Fase 1 — Fundação (Pré-requisito)

Setup do projeto, estrutura base e infraestrutura mínima.

- [ ] Bootstrap Go (go mod, main.go, Makefile)
- [ ] Configuração via env (config)
- [ ] Conexão SQLite + migrações automáticas no boot
- [ ] Entidade Wedding + migração + repositório
- [ ] Router chi com middlewares base (CORS, logging, recovery)
- [ ] Middleware de autenticação JWT (extrai wedding_id dos claims)
- [ ] Middleware TenantResolver (resolve slug → wedding_id no context)
- [ ] Endpoint de login admin (`POST /api/v1/admin/auth`)
- [ ] Health check (`GET /api/v1/health`)
- [ ] Seed do primeiro wedding (manu-rafa) via CLI ou migração
- [ ] Dockerfile
- [ ] Estrutura de resposta padronizada (sucesso e erro)

## Fase 2 — Confirmação de Presença (RSVP) `URGENTE`

Feature principal. Sem isso, os convidados não conseguem confirmar presença.

- [ ] Entidades Invitation + Guest
- [ ] Migrações SQL
- [ ] Repositórios SQLite (scoped por wedding_id)
- [ ] Use case RSVP (buscar por nome no tenant, confirmar, recusar)
- [ ] Use case CRUD de convites (admin, scoped por wedding_id)
- [ ] Use case CRUD de convidados (admin, scoped)
- [ ] Handlers públicos: `POST /w/{slug}/rsvp`, `GET /w/{slug}/rsvp/invitation`
- [ ] Handlers admin: CRUD invitations, CRUD guests, dashboard
- [ ] Testes unitários dos use cases
- [ ] Integração com o frontend (ajustar form action + JS de submit)

## Fase 3 — Lista de Presentes `URGENTE`

Substituir o Casar.com por solução própria com PIX e cartão.

- [ ] Entidades Gift + Payment
- [ ] Migrações SQL
- [ ] Repositórios SQLite (scoped por wedding_id)
- [ ] Integração Mercado Pago (SDK Go) para PIX e cartão
- [ ] Use case: listar presentes (público, por tenant)
- [ ] Use case: iniciar pagamento (público)
- [ ] Use case: webhook de confirmação (Mercado Pago → API)
- [ ] Use case: CRUD de presentes (admin, scoped)
- [ ] Use case: relatório financeiro (admin, scoped)
- [ ] Handlers públicos: `/w/{slug}/gifts`, `/w/{slug}/gifts/{id}/purchase`
- [ ] Handlers admin
- [ ] Testes
- [ ] Integração com o frontend

## Fase 4 — Polimento

- [ ] Seed de dados para desenvolvimento
- [ ] Rate limiting nos endpoints públicos
- [ ] Logs estruturados em produção (JSON)
- [ ] CI/CD básico
- [ ] Deploy (VPS, Fly.io, Railway ou similar)
- [ ] Monitoramento básico (uptime, erros)

## Fase 5 — Plataforma (Futuro)

Evoluções para oferecer o serviço a outros casais.

- [ ] Fluxo de cadastro de novo wedding (self-service ou admin global)
- [ ] Credenciais Mercado Pago por tenant (cada casal recebe na própria conta)
- [ ] Painel super-admin (gestão de todos os weddings)
- [ ] Domínios customizados ou subdomínios por tenant
- [ ] Limites e planos (free, premium)

## Prioridades e Datas

| Fase | Prioridade | Meta |
|------|-----------|------|
| Fase 1 | Bloqueante | Primeira semana |
| Fase 2 | Urgente | Segunda semana |
| Fase 3 | Urgente | Terceira e quarta semana |
| Fase 4 | Importante | Antes do casamento (07.07.2026) |
| Fase 5 | Futuro | Pós-casamento |
