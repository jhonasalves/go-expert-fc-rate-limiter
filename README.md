# Desafio Técnico Go Expert - Rate Limiter

Uma solução configurável e modular para limitar a taxa de requisições em aplicações Go.

## Visão Geral

Este projeto implementa um limitador de taxa que pode ser usado como middleware em um servidor web. Ele oferece:

- Limitação de taxa baseada em IP com limites configuráveis.
- Limitação de taxa baseada em tokens com tempos de expiração personalizados.
- Bloqueio temporário de clientes que excedem os limites.
- Separação clara entre a lógica do limitador de taxa e o middleware.
- Persistência baseada em Redis com suporte para outros mecanismos de armazenamento.

## Instalação

Clone o repositório e navegue até o diretório do projeto:

```bash
git clone https://github.com/jhonasalves/go-expert-fc-rate-limiter.git
cd go-expert-fc-rate-limiter
```

## Início Rápido

1. Configure o arquivo `.env` com os parâmetros desejados.
    ```bash
    RATE_LIMITER_MAX_IP_REQUESTS=10 # Número máximo de requisições por IP
    RATE_LIMITER_MAX_TOKEN_REQUESTS=100 # Número máximo de requisições por Token
    RATE_LIMITER_WINDOW_DURATION=1s # Intervalo de tempo para contar as requisições
    RATE_LIMITER_BLOCK_DURATION=5m  # Tempo de bloqueio após exceder o limite de requisições.

    # Configurações Redis
    REDIS_HOST=localhost
    REDIS_PORT=6379
    REDIS_PASSWORD=
    REDIS_DB=0
    ```

2. Inicie o Redis e servidor usando Docker Compose:
    ```bash
    docker-compose up -d
    ```

3. Acesse o servidor na porta `8080`.

## Configuração

O limitador de taxa suporta as seguintes opções de configuração:

### Valores Padrão
- **Limite baseado em IP**: 10 requisições por segundo.
- **Limite baseado em token**: 100 requisições por segundo.
- **Duração do bloqueio**: 5min (configurável para IPs ou tokens que excedem os limites).
- **Armazenamento**: Redis (via Docker Compose).

### Personalização
Defina os limites e tempos de expiração desejados no arquivo `.env` ou modifique o código conforme necessário.

## Resposta ao Exceder o Limite

Quando um cliente excede o limite de taxa:

### Resposta HTTP
```
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 51
X-Ratelimit-Limit: 2
X-Ratelimit-Remaining: 0
X-Ratelimit-Reset: 1742346764
```

### Resposta JSON
```json
{
  "error": "rate_limit_exceeded",
  "message": "you have reached the maximum number of requests or actions allowed within a certain time frame",
  "limit": 2,
  "remaining": 0,
  "reset_after": 51
}
```

## Armazenamento

### Armazenamento Redis (Padrão)
O Redis é usado para persistência. Para iniciar o Redis:

```bash
docker-compose up -d
```

### Armazenamento Alternativo
O projeto foi projetado para suportar outros mecanismos de armazenamento. Implemente a interface necessária para integrar um novo backend.

## Exemplos

### Limitação Baseada em IP
- **Configuração**: 5 requisições por segundo.
- **Exemplo**: O IP `192.168.1.1` envia 6 requisições em 1 segundo. A 6ª requisição é bloqueada.

### Limitação Baseada em Token
- **Configuração**: Token `abc123` com limite de 10 requisições por segundo.
- **Exemplo**: O token envia 11 requisições em 1 segundo. A 11ª requisição é bloqueada.

### Executando Testes

```bash
go test ./...
```