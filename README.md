# Rate Limiter com Redis e Janela Fixa

Este projeto implementa um **Rate Limiter** em Go utilizando Redis para armazenamento de estado. Ele controla o número de requisições por segundo com base no endereço IP e/ou no token de acesso enviado na requisição.

## Como Funciona

1. **Janela Fixa (Fixed Window)**:
   - A lógica do rate limiter utiliza o modelo de **janela fixa**.
   - Uma nova janela de tempo é iniciada assim que uma requisição é recebida de um IP ou token cuja janela ainda não foi registrada.
   - Durante a janela de tempo, o sistema monitora a quantidade de requisições e bloqueia requisições excedentes.

2. **Prioridade do Token**:
   - Caso a requisição contenha um **token** (especificado no cabeçalho `API_KEY: <TOKEN>`), as configurações do token terão prioridade sobre as configurações de IP.
   - Se não houver configurações específicas para o token, serão utilizadas configurações padrão definidas pelas variáveis `MAX_REQUESTS_PER_SECOND_TOKEN` e `RATE_INTERVAL_TOKEN`.

3. **Tempo de Bloqueio**:
   - Quando o limite de requisições é atingido, o IP ou token será **bloqueado** por um período de tempo definido:
     - Para IPs: configurado com `BLOCK_DURATION_IP`.
     - Para tokens: configurado com `BLOCK_DURATION_TOKEN`.
   - Durante o bloqueio, todas as requisições retornarão o código HTTP `429` com a mensagem:
     ```
     you have reached the maximum number of requests or actions allowed within a certain time frame
     ```

## Configuração

O rate limiter pode ser configurado no arquivo `.env`.

### Exemplo de configuração
Abaixo está um exemplo de configuração:

```
MAX_REQUESTS_PER_SECOND_IP=10
MAX_REQUESTS_PER_SECOND_TOKEN=20
BLOCK_DURATION_IP=10s
BLOCK_DURATION_TOKEN=60s
RATE_INTERVAL_IP=1
RATE_INTERVAL_TOKEN=30
REDIS_ADDRESS=localhost:6379
REDIS_PASSWORD=
WEB_SERVER_PORT=8080
TOKENS=JOAO123:10/1,CARLOS456:30/2
```

### Explicação das Variáveis

**Limites por IP**
- `MAX_REQUESTS_PER_SECOND_IP`: Número máximo de requisições permitidas por intervalo de tempo para um único IP.  
- `RATE_INTERVAL_IP`: Intervalo de tempo (em segundos) da janela fixa para IPs.  
- `BLOCK_DURATION_IP`: Tempo de bloqueio (em segundos) para um IP que ultrapassou o limite.

**Limites por Token**
- `MAX_REQUESTS_PER_SECOND_TOKEN`: Limite padrão de requisições por intervalo de tempo para tokens que não possuem configurações específicas.  
- `RATE_INTERVAL_TOKEN`: Intervalo padrão (em segundos) da janela fixa para tokens sem configurações específicas.  
- `BLOCK_DURATION_TOKEN`: Tempo de bloqueio (em segundos) para um token que ultrapassou o limite.

**Configurações de Tokens Específicos** 
- `TOKENS`: Lista de tokens com suas configurações específicas, no formato `<token>:<limite>/<intervalo>`. Exemplo:
    - `JOAO123:10/1`: O token `JOAO123` permite 10 requisições por segundo.
    - `CARLOS456:20/2`: O token `CARLOS456` permite 20 requisições a cada 2 segundos.
## Mecanismo de Persistência
O rate limiter foi projetado para suportar mecanismos de persistência através do Redis.

## Testes Automatizados
O sistema possui testes automatizados que validam o funcionamento do rate limiter. O teste principal está localizado no arquivo:

```
/middleware/middleware_test.go
```

## Como Usar

### Execução Local

1.  **Clone o repositório (se aplicável).**
2.  **Navegue até o diretório do projeto.**
3.  **Compile o código Go:**
    ```bash
    go run ./cmd/ratelimiter/ratelimiter.go
    ```
4.  **Testes de requisição**
    Para testes de requisição por IP, executar o seguinte arquivo em outro terminal:
    ```bash
    go run ./test/ip/requestIP.go
    ```

    Para testes de requisição por token, executar o seguinte arquivo em outro terminal:
    ```bash
    go run ./test/token/requestToken.go
    ```

    