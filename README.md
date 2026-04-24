# wterm

Servidor de terminal remoto via **WebSocket** escrito em **Go**, usando **PTY (pseudo-terminal)** para rodar um shell real de forma interativa no navegador.

> **Atenção (segurança):** este projeto expõe um shell remoto. Use somente em ambientes controlados. **Troque a senha e o segredo do servidor antes de rodar em produção**.

## Visão geral

- Backend em Go com:
  - Endpoint `POST /auth` para gerar um **token** e uma **assinatura HMAC**
  - Endpoint `GET /` que faz upgrade para **WebSocket** e liga o socket a um **PTY**
- Frontend simples (xterm.js) em `view/index.html`.

## Como funciona (alto nível)

1. O cliente envia a senha para `POST /auth`.
2. O servidor retorna `{ token, signature }`.
3. O cliente abre `wss://.../?token=...&signature=...`.
4. O servidor valida HMAC e, se OK, inicia um shell (via PTY) e faz o “pipe” entre PTY ⇄ WebSocket.

## Requisitos

- Go (para rodar localmente)
- Ou Docker / Docker Compose (recomendado)

## Configuração de segurança (obrigatório)

No arquivo `main.go`, altere:

- `const serverSecret = "CHAVE-MUITO-SECRETA-ALTERAR-AQUI"`
- `const password = "123"`

Recomendação: use valores longos, aleatórios e não versionados (env vars / secrets).

## Rodando localmente (Go)

```bash
go mod download
go run . -port=9090
```

Depois conecte com um cliente WebSocket (ou use o `view/index.html` apontando para seu host).

## Rodando com Docker

Build:

```bash
docker build -t wterm .
```

Run:

```bash
docker run --rm -p 9090:9090 \
  -e SHELL=/bin/bash \
  -e HOME=/root \
  wterm
```

## Rodando com Docker Compose

O repositório inclui um `docker-compose.yml`. Para subir:

```bash
docker compose up -d
```

Por padrão, ele:

- expõe `9090:9090`
- monta `./workspace` em `/root` (diretório HOME dentro do container)
- monta `/var/run/docker.sock` (use com cuidado)

## Endpoints

- `POST /auth` (form urlencoded)
  - body: `password=<sua_senha>`
  - response: JSON com `token` e `signature`
- `GET /` (WebSocket)
  - query: `token` e `signature`

## Frontend (xterm.js)

O arquivo `view/index.html` usa xterm.js via CDN e implementa um fluxo simples de login:

1. pede a senha
2. chama `/auth`
3. abre o WebSocket com o token assinado

> Se você for rodar localmente, ajuste as URLs hardcoded no `view/index.html` para apontar para o seu servidor.

## Licença

Sem licença definida no momento. Se quiser, posso adicionar uma licença (ex.: MIT) e atualizar o README para refletir.
