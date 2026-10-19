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
