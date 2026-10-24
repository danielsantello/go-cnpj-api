# Índices do MySQL

A API mantém o catálogo de índices necessários em
`internal/storage/mysql/v1/indexes.go`.

Para criar os índices, execute na raiz do projeto:

```bash
clear
set -a
source .env
set +a
time go run ./cmd/cnpj-api database create-indexes
```

O arquivo `.env` deve conter a configuração de conexão com o MySQL.

O comando verifica os índices existentes e cria os ausentes. Execuções
posteriores não reconstroem os índices já existentes.

## Tempos observados

Referência obtida com MySQL 8.4.11 em container Docker, usando a base
da publicação 2026/08. Havia aproximadamente 848 GB livres em disco
na primeira criação.

| Tabela / colunas indexadas | Tempo observado |
|---|---:|
| `establishments`: `cnpj` e `cnpj_root` | 5m09s |
| `companies`: `cnpj_root` | 2m48s |
| Tabelas pequenas de domínio | Menos de 1s cada |
| `simple_tax_options`: `cnpj_root` | 1m54s |
| `partners`: `cnpj_root` | 1m11s |
| Primeira criação dos 11 índices | 11m02,542s |
| `companies`: `(legal_name, cnpj_root)` | 5m09,641s |
| Total acumulado dos 12 índices | Aproximadamente 16m12s |
| `establishments`: `(trade_name, cnpj)` | 5m15s |
| Total acumulado dos 13 índices | Aproximadamente 21m27s |

A criação do índice `(trade_name, cnpj)` foi medida pelos logs:
de 20:06:18 a 20:11:33 em 19/10/2026, com resultado
`created=1 existing=12`.

O levantamento anterior usou `time` para medir as execuções completas
e os horários dos logs para os intervalos por tabela. O total dos
13 índices é uma soma aproximada de execuções distintas.

Esses tempos são referências observadas, não garantias. Eles variam
conforme processador, disco, memória, configuração do MySQL, carga
concorrente, quantidade de registros e tamanho dos textos indexados.

## Execuções sem índices pendentes

| Índices existentes | Resultado | Tempo total com `time` |
|---|---|---:|
| 11 | `created=0 existing=11` | 0,126s |
| 12 | `created=0 existing=12` | 0,596s |

Esses tempos incluem inicialização do programa, conexão com o MySQL,
validação do schema e consulta dos índices existentes.

## Dimensão da base utilizada

Estimativas obtidas de `information_schema.tables`:

| Tabela | Registros estimados | Tamanho dos dados |
|---|---:|---:|
| `establishments` | 64.590.538 | 12,99 GB |
| `companies` | 62.139.658 | 5,71 GB |
| `simple_tax_options` | 48.215.694 | 2,25 GB |
| `partners` | 26.232.166 | 2,67 GB |

## Índice de busca por sócio

Índice `idx_partners_name_document_cnpj_root`, nas colunas
`(partner_name, partner_document, cnpj_root)`.

Criado em 20/10/2026, no mesmo ambiente de referência:

| Medição | Tempo observado |
|---|---:|
| Criação pelos logs, de 11:04:01 a 11:06:00 | 1m59s |
| Execução completa com `time go run` | 1m59,686s |

Resultado: `created=1 existing=13`, totalizando 14 índices no catálogo.

O total acumulado aproximado das criações passa a 23m27s,
somando execuções distintas.

## Índice de seleção do estabelecimento representante

Índice `idx_establishments_root_head_office_cnpj`, nas colunas
`(cnpj_root, head_office_branch_indicator, cnpj)`. Total do catálogo: 15 índices.

Criado em 23/10/2026 na mesma publicação de referência:

| Medição | Tempo observado |
|---|---:|
| Criação pelos logs, de 20:11:12 a 20:41:27 | 30m15s |
| Execução completa com `time go run` | 30m16,418s |

Resultado: `created=1 existing=14`. O índice cobre a seleção da matriz ou,
na ausência dela, do menor CNPJ. A expressão de prioridade ainda pode
exigir ordenação. Na leitura para indexação de sócios, a subconsulta usa
`FORCE INDEX (idx_establishments_root_head_office_cnpj)`, pois o plano
observado continuava escolhendo o índice simples após a criação.

A CTE que seleciona as empresas do lote usa `SELECT STRAIGHT_JOIN` para
impedir que os `EXISTS` sejam convertidos em um semijoin que percorra
milhões de sócios antes do limite. Não há alteração nos registros da origem.

Na amostra inicial do CNPJ básico `00000000`, com 41 vínculos:

| Operação | Antes do uso do índice composto | Depois |
|---|---:|---:|
| Seleção isolada do CNPJ representante | 1.015,662 ms | 11,672 ms |
| Leituras de lotes de 1 ou 8 vínculos | 804,967–1.128,493 ms | 7,392–15,457 ms |

São observações de execuções sequenciais, sem controle de cache ou
concorrência. O teste completo também faz consultas de conferência;
sua duração não representa apenas o custo da leitura para indexação.
Esses tempos não estimam a duração de uma carga completa.
