# Endpoints

Endereço padrão: `http://localhost:8080`.

## Disponibilidade

`GET /health`

Verifica a disponibilidade da API e a conexão com o MySQL.

```bash
clear
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
status `unhealthy`. As buscas específicas podem continuar pelo MySQL mesmo
com o health indicando indisponibilidade do Manticore. O health verifica a
conexão, não a existência nem a completude das tabelas de busca.

## Consulta por CNPJ

`GET /v1/companies/{cnpj}`

Retorna dados da empresa, do estabelecimento consultado, do Simples
Nacional e MEI, quando disponíveis, e dos sócios.

```bash
clear
curl -sS 'http://localhost:8080/v1/companies/00523856000141'
```

### Incluir os estabelecimentos da empresa

```bash
clear
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

Busca termos somente na razão social quando o Manticore está disponível.
Quando desabilitado ou indisponível, busca pelo início da razão social no MySQL.

```bash
clear
curl -sS --get 'http://localhost:8080/v1/companies' \
  --data-urlencode 'name=SANTELLO' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

Cada item contém `cnpj`, `corporate_name`, `trade_name`, `municipality`,
`state` e `registration_status`. O CNPJ retornado é preferencialmente
o da matriz; na ausência dela, o menor CNPJ disponível para a empresa.

`trade_name` é o nome fantasia do estabelecimento do CNPJ completo
retornado. O campo está sempre presente: contém uma string ou `null`
quando o valor é nulo na origem. Uma string vazia é preservada.

No caminho Manticore, o campo é obtido no mesmo lote de enriquecimento
no MySQL e também retorna `null` quando o CNPJ não é encontrado.
No fallback MySQL, é lido diretamente do estabelecimento selecionado.

A inclusão do campo não altera os campos pesquisados, a ordenação,
a paginação ou as regras de fallback. Não requer recarga do Manticore.

Os resultados ficam em `data.companies.items`. No MySQL, são ordenados por
razão social e CNPJ básico; no Manticore, pelo ID do documento da empresa.

## Busca por nome fantasia

`GET /v1/establishments`

Busca termos somente no nome fantasia de cada estabelecimento quando o
Manticore está disponível. Quando desabilitado ou indisponível, busca pelo
início do nome fantasia no MySQL.

```bash
clear
curl -sS --get 'http://localhost:8080/v1/establishments' \
  --data-urlencode 'name=SANTELLO' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

Cada item contém `cnpj`, `trade_name`, `corporate_name`, `municipality`, `state` e `registration_status`.

Os resultados ficam em `data.establishments.items`. No MySQL, são ordenados
por nome fantasia e CNPJ; no Manticore, pelo ID do documento do estabelecimento.

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

Nas buscas com Manticore, o CNPJ é o armazenado no documento correspondente.
Os novos campos são consultados no MySQL em um único lote com os
CNPJs da página, preservando itens, ordem, correspondências e paginação.
Páginas vazias não executam essa consulta adicional.

Se um CNPJ não for encontrado no MySQL, o item é mantido com os novos
campos nulos. Uma falha na consulta de enriquecimento retorna HTTP 500.
MySQL e Manticore devem apontar para a mesma publicação.

O enriquecimento não altera a estrutura dos documentos de busca. A seleção
automática exige índices específicos por estabelecimento e por vínculo de sócio;
o índice agregado por empresa continua atendendo razão social e `/v1/search`.

## Parâmetros das buscas por nome

Válidos para as buscas por razão social e nome fantasia:

| Parâmetro | Obrigatório | Descrição |
|---|---|---|
| `name` | Sim | Nome entre 3 e 255 caracteres; termos no Manticore ou prefixo no MySQL |
| `page` | Não | Página solicitada; padrão: 1 |
| `page_size` | Não | Itens por página, de 1 a 100; padrão: 20 |

Espaços nas extremidades de `name` são removidos.
No MySQL, `%` e `_` são tratados literalmente. No Manticore, a pontuação
é separadora e os operadores não são expostos. Todos os termos são
obrigatórios, em qualquer ordem e posição no campo correspondente.
Um texto composto somente por pontuação retorna lista vazia no Manticore.

A seleção do mecanismo é independente em cada endpoint. Não há fallback
por resultado vazio, entrada inválida, `SEARCH_TOO_BROAD`, erro SQL inesperado,
falha de enriquecimento no MySQL ou cancelamento da requisição. Falhas de
conexão, de leitura/escrita por timeout e ausência da tabela de busca permitem
fallback. Uma tabela existente, porém incompleta, não aciona fallback.

Cada resposta usa somente um mecanismo: itens e total não são misturados
entre Manticore e MySQL. O fallback refaz a busca inteira pelo prefixo.
A ordem e o universo de resultados podem mudar entre requisições se o
mecanismo disponível mudar. Consulte [operação das buscas](search.md).

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

Busca termos no nome do sócio usando Manticore, com fallback por prefixo
no MySQL quando desabilitado ou indisponível:

```bash
clear
curl -sS --get 'http://localhost:8080/v1/partners' \
  --data-urlencode 'name=SANTELLO' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

O parâmetro opcional `document` acrescenta uma condição AND por igualdade
ao documento armazenado. O nome mantém a busca por termos ou prefixo,
conforme o mecanismo. A máscara é literal; `*` não funciona como curinga:

```bash
clear
curl -sS --get 'http://localhost:8080/v1/partners' \
  --data-urlencode 'name=SANTELLO PRA NIEDENS' \
  --data-urlencode 'document=***111769**'
```

| Parâmetro | Obrigatório | Descrição |
|---|---|---|
| `name` | Sim | Nome entre 3 e 255 caracteres; termos no Manticore ou prefixo no MySQL, inclusive com documento |
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
A unidade do resultado é o vínculo entre sócio e empresa: vínculos distintos
com o mesmo nome e documento não são fundidos. O filtro de documento é
aplicado antes da contagem, do limite de 1000 e da paginação.
Não há esse filtro em `/v1/search`.

A resposta inclui `data.partners.pagination`, no mesmo formato das
demais buscas. Sem resultados, retorna HTTP 200 com `items: []`.

Parâmetros inválidos retornam HTTP 400 com código
`INVALID_QUERY_PARAMETERS`.

## Busca ampla — Manticore

`GET /v1/search`

Busca termos na razão social, nos nomes fantasia e nos nomes dos sócios.
Requer Manticore habilitado e dados indexados.

```bash
clear
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
do documento, com paginação em `data.companies.pagination`. Os novos índices
usam IDs sequenciais independentes do valor textual do CNPJ.

Sem resultados, retorna HTTP 200 com `items: []`.
Com Manticore desabilitado, retorna HTTP 503 com código `SEARCH_DISABLED`.
Esse endpoint mantém o contrato de busca ampla e não utiliza o fallback dos
três endpoints específicos; falhas de consulta retornam HTTP 500.

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

## Medição local do enriquecimento — comportamento anterior

Esta medição antecede a seleção automática: as três buscas específicas
utilizavam prefixo no MySQL. Os universos das buscas por termos e por
prefixo são diferentes, portanto seus tempos não são comparáveis diretamente.

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

## Consumo pelo Explorer

A busca geral do Explorer pode chamar `/v1/companies`, `/v1/establishments`
e `/v1/partners`, exibindo três seções com paginações independentes.
Cada chamada filtra, conta e pagina sua própria entidade. Não se deve somar
os totais nem aplicar `distinct` aos itens de páginas diferentes para criar
uma paginação única. `/v1/search` continua sendo uma operação separada,
agregada por empresa e com `matched_fields`.
