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

O campo `search` representa o Manticore Search, uma integração opcional.
Quando desabilitado, retorna `disabled`; quando habilitado, retorna
`available` ou `unavailable`, conforme a conexão. Se qualquer serviço
habilitado estiver indisponível, o endpoint retorna HTTP 503 com
status `unhealthy`. As buscas por prefixo utilizam o MySQL.

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

Cada item contém `cnpj`, `corporate_name`, `municipality`, `state` e `registration_status`. O CNPJ retornado é
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

Cada item contém `cnpj`, `trade_name`, `corporate_name`, `municipality`, `state` e `registration_status`.

Os resultados ficam em `data.establishments.items`, ordenados
por nome fantasia e CNPJ.

## Município, UF e situação cadastral nos resumos

As buscas por razão social, nome fantasia, sócio e termos incluem
estes campos em cada item:

| Campo | Conteúdo |
|---|---|
| `municipality` | Objeto com `code` e `description`, ou `null` |
| `state` | Sigla da UF, ou `null` |
| `registration_status` | Objeto com `code` e `description`, ou `null` |

Os dados correspondem ao estabelecimento do CNPJ completo retornado.
Uma descrição indisponível pode ser `null`, mesmo quando há um código.

Nas buscas por razão social e sócio, o estabelecimento escolhido é
preferencialmente a matriz; na ausência dela, o de menor CNPJ.
Na busca por nome fantasia, os dados são do estabelecimento encontrado.

Na busca por termos, o CNPJ é o armazenado no documento do Manticore.
Os novos campos são consultados no MySQL em um único lote com os
CNPJs da página, preservando itens, ordem, correspondências e paginação.
Páginas vazias não executam essa consulta adicional.

Se um CNPJ não for encontrado no MySQL, o item é mantido com os novos
campos nulos. Uma falha na consulta de enriquecimento retorna HTTP 500.
MySQL e Manticore devem apontar para a mesma publicação.

Essa melhoria não exige alterar a estrutura nem recarregar o Manticore.

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
| 422 | `SEARCH_TOO_BROAD` | Busca com mais de 1000 resultados |
| 500 | `INTERNAL_ERROR` | Erro interno durante a consulta |
| 503 | `SEARCH_DISABLED` | Busca ampla com Manticore desabilitado |

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

## Busca por sócio

`GET /v1/partners`

Sem `document`, busca pelo início do nome do sócio:

```bash
curl -sS --get 'http://localhost:8080/v1/partners' \
  --data-urlencode 'name=SANTELLO' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

Com `document`, busca por nome completo e documento por igualdade:

```bash
curl -sS --get 'http://localhost:8080/v1/partners' \
  --data-urlencode 'name=SANTELLO PRA NIEDENS' \
  --data-urlencode 'document=***111769**'
```

| Parâmetro | Obrigatório | Descrição |
|---|---|---|
| `name` | Sim | Nome entre 3 e 255 caracteres; prefixo sem documento, nome completo com documento |
| `document` | Não | Documento como consta na base, incluindo a máscara; entre 1 e 20 caracteres |
| `page` | Não | Página solicitada; padrão: 1 |
| `page_size` | Não | Itens por página, de 1 a 100; padrão: 20 |

Espaços nas extremidades do nome e do documento são removidos.
Na busca por prefixo, `%` e `_` são tratados literalmente.

Os resultados ficam em `data.partners.items`. Cada item contém:

- `name`: nome do sócio.
- `document`: documento conforme a base, podendo ser nulo.
- `cnpj`: CNPJ da empresa vinculada, preferencialmente o da matriz.
- `corporate_name`: razão social da empresa.
- `municipality`: município do estabelecimento retornado, com código e descrição, ou nulo.
- `state`: UF do estabelecimento retornado, ou nula.
- `registration_status`: situação cadastral do estabelecimento retornado, com código e descrição, ou nula.

Um sócio pode aparecer vinculado a várias empresas. O documento
mascarado ajuda a distinguir homônimos, mas não garante identidade.

A resposta inclui `data.partners.pagination`, no mesmo formato das
demais buscas. Sem resultados, retorna HTTP 200 com `items: []`.

Parâmetros inválidos retornam HTTP 400 com código
`INVALID_QUERY_PARAMETERS`.

## Busca ampla — Manticore

`GET /v1/search`

Busca termos na razão social, nos nomes fantasia e nos nomes dos sócios.
Requer Manticore habilitado e dados indexados.

```bash
curl -sS --get 'http://localhost:8080/v1/search' \
  --data-urlencode 'name=banco brasil' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

| Parâmetro | Obrigatório | Descrição |
|---|---|---|
| `name` | Sim | Texto entre 3 e 255 caracteres |
| `page` | Não | Página solicitada; padrão: 1 |
| `page_size` | Não | Itens por página, de 1 a 100; padrão: 20 |

Todos os termos são obrigatórios, em qualquer ordem. Eles podem aparecer
em campos diferentes do mesmo documento de empresa.

A busca considera palavras em qualquer posição dos campos, com
normalização de maiúsculas, minúsculas e acentos. Não é uma busca
por qualquer trecho dentro de uma palavra.

Pontuação é tratada como separador; operadores de consulta do Manticore
não são expostos. Texto composto somente por pontuação retorna lista vazia.

Cada empresa aparece uma vez, com `basic_cnpj`, `cnpj`, `corporate_name`, `matched_fields`, `municipality`, `state` e `registration_status`.

O campo `matched_fields` indica onde pelo menos um dos termos foi
encontrado: `corporate_name` (razão social), `trade_names` (nomes fantasia)
ou `partner_names` (nomes dos sócios).
Os resultados ficam em `data.companies.items`, ordenados pelo identificador
derivado do CNPJ básico, com paginação em `data.companies.pagination`.

Sem resultados, retorna HTTP 200 com `items: []`.
Com Manticore desabilitado, retorna HTTP 503 com código `SEARCH_DISABLED`.

### Estado atual

A tabela consultada é definida por `CNPJ_API_MANTICORE_TABLE`.
A troca exige reiniciar a API após concluir e validar a carga.

### Limite de resultados

A busca permite até 1000 empresas por consulta, com paginação.

Acima desse limite, retorna HTTP 422 com código `SEARCH_TOO_BROAD`,
sem itens nem paginação. Refine a busca acrescentando termos.

## Limite comum às buscas por nome

As buscas por razão social, nome fantasia, sócio e a busca ampla
permitem até 1000 resultados por consulta.

Até esse limite, retornam itens e paginação normalmente.
Acima dele, retornam HTTP 422 com código `SEARCH_TOO_BROAD`,
sem itens nem paginação. Refine o nome ou acrescente termos
na busca ampla.

Na busca por sócio, o limite também se aplica quando o documento
é informado.

Essa regra não altera a consulta direta por CNPJ nem a listagem
de estabelecimentos incluída nos detalhes de uma empresa.

## Medição local do enriquecimento

Medição realizada em 22/10/2026, com a consulta `SANTELLO`, página 1,
MySQL 8.4.11 e Manticore 29.9.0 locais, na publicação 2026/08.

Cada combinação teve um aquecimento e cinco requisições sequenciais.
Os valores abaixo são medianas da latência HTTP, em milissegundos.
Todas as requisições retornaram HTTP 200.

| Busca | `page_size` | Antes | Depois | Diferença |
|---|---:|---:|---:|---:|
| Razão social | 20 | 3,567 | 3,319 | -0,248 |
| Razão social | 100 | 3,371 | 3,156 | -0,215 |
| Nome fantasia | 20 | 3,048 | 2,711 | -0,337 |
| Nome fantasia | 100 | 2,610 | 2,746 | +0,136 |
| Sócios | 20 | 2,933 | 2,663 | -0,270 |
| Sócios | 100 | 2,967 | 2,599 | -0,368 |
| Termos | 20 | 6,415 | 8,177 | +1,762 |
| Termos | 100 | 8,874 | 13,476 | +4,602 |

A consulta retornou 30 empresas, 12 estabelecimentos, um vínculo de
sócio e 321 empresas na busca por termos. O tamanho solicitado não
representa necessariamente a quantidade de itens retornados.

As pequenas diferenças nas buscas MySQL não comprovam ganho de
desempenho. Na busca por termos, o acréscimo observado foi de
aproximadamente 1,8 ms para 20 itens e 4,6 ms para 100 itens.

Esta amostra não mede concorrência, carga sustentada nem o tempo
isolado das consultas SQL.
