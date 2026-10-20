# Endpoints

Endereço padrão: `http://localhost:8080`.

## Disponibilidade

`GET /health`

Verifica a disponibilidade da API e a conexão com o MySQL.

```bash
curl -i 'http://localhost:8080/health'
```

Com o MySQL disponível, retorna HTTP 200:

```json
{
  "status": "healthy",
  "services": {
    "mysql": {
      "status": "available"
    },
    "search": {
      "status": "disabled"
    }
  }
}
```

Se a verificação do MySQL falhar, retorna HTTP 503, com status
`unhealthy` e MySQL `unavailable`.

O campo `search` representa o serviço adicional de busca, atualmente
desabilitado. As buscas por prefixo funcionam pelo MySQL.

## Consulta por CNPJ

`GET /v1/companies/{cnpj}`

Retorna dados da empresa, do estabelecimento consultado, do Simples
Nacional e MEI, quando disponíveis, e dos sócios.

```bash
curl -sS 'http://localhost:8080/v1/companies/00523856000141'
```

### Incluir os estabelecimentos da empresa

```bash
curl -sS --get \
  'http://localhost:8080/v1/companies/00523856000141' \
  --data-urlencode 'include=establishments' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

| Parâmetro | Obrigatório | Descrição |
|---|---|---|
| `include` | Não | Use `establishments` para incluir a lista de estabelecimentos da empresa |
| `page` | Não | Página da lista de estabelecimentos; padrão: 1 |
| `page_size` | Não | Itens por página, de 1 a 100; padrão: 20 |

A resposta contém `data.company`, `data.establishment`,
`data.simple_tax` e `data.partners`. Com `include=establishments`,
inclui também `data.establishments`, com itens e paginação.

O CNPJ é convertido para maiúsculas, caracteres diferentes de letras
e números são removidos e valores menores que 14 caracteres recebem
zeros à esquerda. Não há validação de dígitos verificadores.

## Busca por razão social

`GET /v1/companies`

Retorna empresas cuja razão social começa com o nome informado.

```bash
curl -sS --get 'http://localhost:8080/v1/companies' \
  --data-urlencode 'name=SANTELLO' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

Cada item contém `cnpj` e `corporate_name`. O CNPJ retornado é
preferencialmente o da matriz; na ausência dela, o menor CNPJ
disponível para a empresa.

Os resultados ficam em `data.companies.items`, ordenados por
razão social e CNPJ básico da empresa.

## Busca por nome fantasia

`GET /v1/establishments`

Retorna estabelecimentos cujo nome fantasia começa com o nome informado.

```bash
curl -sS --get 'http://localhost:8080/v1/establishments' \
  --data-urlencode 'name=SANTELLO' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

Cada item contém `cnpj`, `trade_name` e `corporate_name`.

Os resultados ficam em `data.establishments.items`, ordenados
por nome fantasia e CNPJ.

## Parâmetros das buscas por nome

Válidos para as buscas por razão social e nome fantasia:

| Parâmetro | Obrigatório | Descrição |
|---|---|---|
| `name` | Sim | Prefixo do nome, entre 3 e 255 caracteres |
| `page` | Não | Página solicitada; padrão: 1 |
| `page_size` | Não | Itens por página, de 1 a 100; padrão: 20 |

Espaços nas extremidades de `name` são removidos.
Os caracteres `%` e `_` são tratados literalmente.

As buscas são por início do nome, não por palavras em qualquer posição.

### Paginação

As listas paginadas incluem:

```json
{
  "page": 1,
  "page_size": 5,
  "total_items": 12,
  "total_pages": 3
}
```

Nas buscas por nome, esse objeto fica em
`data.companies.pagination` ou `data.establishments.pagination`.

Uma busca sem resultados retorna HTTP 200 com `items: []`.
Uma página além dos resultados disponíveis também retorna uma lista vazia.

## Erros

| HTTP | Código | Situação |
|---|---|---|
| 400 | `INVALID_QUERY_PARAMETERS` | Nome, inclusão ou paginação inválidos |
| 400 | `INVALID_CNPJ` | CNPJ com mais de 14 caracteres após a normalização |
| 404 | `COMPANY_NOT_FOUND` | Nenhuma empresa encontrada para o CNPJ |
| 500 | `INTERNAL_ERROR` | Erro interno durante a consulta |

Os erros das consultas usam o formato:

```json
{
  "error": {
    "code": "INVALID_QUERY_PARAMETERS",
    "message": "One or more query parameters are invalid.",
    "details": [
      {
        "field": "page",
        "reason": "page must be a positive integer."
      }
    ],
    "request_id": "identificador-da-requisicao"
  }
}
```

O identificador também é retornado no cabeçalho `X-Request-ID`.

## Índices

Consulte [Índices do MySQL](indexes.md) para o comando de criação
e os tempos observados.
