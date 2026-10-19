# Endpoints

## Busca por nome fantasia

`GET /v1/establishments`

Retorna estabelecimentos cujo nome fantasia começa com o valor informado.
Cada item inclui CNPJ, nome fantasia e razão social da empresa.

### Parâmetros

| Parâmetro | Obrigatório | Descrição |
|---|---|---|
| `name` | Sim | Prefixo do nome fantasia, entre 3 e 255 caracteres |
| `page` | Não | Página solicitada; padrão: 1 |
| `page_size` | Não | Itens por página, entre 1 e 100; padrão: 20 |

Espaços nas extremidades de `name` são removidos antes da validação.
Os caracteres `%` e `_` são tratados literalmente.

Os resultados são ordenados por nome fantasia e CNPJ.

### Exemplo

```bash
clear
curl -sS --get 'http://localhost:8080/v1/establishments' \
  --data-urlencode 'name=SANTELLO' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

### Estrutura da resposta

Exemplo ilustrativo:

```json
{
  "data": {
    "establishments": {
      "items": [
        {
          "cnpj": "12345678000190",
          "trade_name": "LOJA EXEMPLO",
          "corporate_name": "EMPRESA EXEMPLO LTDA"
        }
      ],
      "pagination": {
        "page": 1,
        "page_size": 20,
        "total_items": 1,
        "total_pages": 1
      }
    }
  }
}
```

Uma busca sem resultados retorna HTTP 200 com `items: []`.
Uma página além dos resultados disponíveis também retorna uma lista vazia.

### Erros

| HTTP | Código | Situação |
|---|---|---|
| 400 | `INVALID_QUERY_PARAMETERS` | Nome ou paginação inválidos |
| 500 | `INTERNAL_ERROR` | Erro interno durante a busca |

### Índice utilizado

`idx_establishments_trade_name_cnpj`, nas colunas `(trade_name, cnpj)`.

Consulte [Índices do MySQL](indexes.md) para instruções de criação
e tempos observados.
