# -------- BUILD STAGE --------
FROM golang:1.25 AS builder

WORKDIR /app

# Copia go.mod e go.sum primeiro (cache melhor)
COPY go.mod go.sum ./
RUN go mod download

# Copia o resto do código
COPY . .

# Compila binário estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

# -------- RUNTIME STAGE --------
FROM debian:bookworm-slim

WORKDIR /app

# Instala bash (necessário para o teu shell)
RUN apt-get update && apt-get install -y \
    bash \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copia o binário do build
COPY --from=builder /app/server .

# Variáveis de ambiente
ENV SHELL=/bin/bash
ENV HOME=/root

# Porta
EXPOSE 9090

# Comando para rodar
CMD ["./server", "-port=9090"]