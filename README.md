# Clima por CEP — Google Cloud Run

Sistema em Go que recebe um CEP, identifica a cidade via [ViaCEP](https://viacep.com.br) e retorna a temperatura atual (Celsius, Fahrenheit e Kelvin) via [WeatherAPI](https://www.weatherapi.com).

## URL no Cloud Run

**https://clima-cep-340522919999.us-central1.run.app**

Exemplo:

```
GET https://clima-cep-340522919999.us-central1.run.app/weather/01001000
```

## API

| Rota | Exemplo |
|------|---------|
| `GET /weather/{cep}` | `/weather/01001000` |
| `GET /?cep={cep}` | `/?cep=01001000` |

### Respostas

| Cenário | Status | Corpo |
|---------|--------|-------|
| Sucesso | `200` | `{"city":"São Paulo","state":"São Paulo","temp_C":28.5,"temp_F":83.3,"temp_K":301.5}` |
| CEP com formato inválido (≠ 8 dígitos, letras, hífen) | `422` | `invalid zipcode` |
| CEP inexistente | `404` | `can not find zipcode` |
| Falha em serviço externo | `500` | `internal server error` |

Conversões: `F = C × 1.8 + 32` e `K = C + 273` (valores arredondados em 2 casas).

## Pré-requisitos

- Docker
- Uma chave gratuita da WeatherAPI: https://www.weatherapi.com/signup.aspx

```bash
cp .env.example .env
# edite .env e informe WEATHER_API_KEY
```

## Rodando a aplicação localmente via Docker

Com Docker Compose:

```bash
docker compose up --build app
curl http://localhost:8080/weather/01001000
```

Ou apenas com Docker:

```bash
docker build -t clima-cep .
docker run --rm -p 8080:8080 -e WEATHER_API_KEY=sua_chave clima-cep
curl http://localhost:8080/weather/01001000
```

## Rodando os testes

Os testes não acessam a internet nem precisam de chave de API (as APIs externas são simuladas com `httptest`). Eles cobrem as conversões de temperatura, os clientes ViaCEP/WeatherAPI e os cenários 200, 404, 422 e 500 do handler.

Via Docker:

```bash
docker build --target test .
# ou
docker compose --profile test run --rm --build test
```

Com Go instalado (1.26+):

```bash
go test ./...
```

## Deploy no Google Cloud Run

```bash
gcloud auth login
gcloud config set project SEU_PROJETO

gcloud run deploy clima-cep \
  --source . \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars WEATHER_API_KEY=sua_chave
```

O comando retorna a URL do serviço; atualize a seção **URL no Cloud Run** acima.

## Estrutura

```
cmd/server/            ponto de entrada (lê PORT e WEATHER_API_KEY)
internal/handler/      handler HTTP, validação do CEP e mapeamento de erros
internal/viacep/       cliente ViaCEP
internal/weatherapi/   cliente WeatherAPI
internal/temperature/  conversões de temperatura
```
