# go-cnpj-api

API em Go para consulta de dados públicos de CNPJ da Receita Federal.

Permite consultar uma empresa pelo CNPJ e pesquisar por início da razão
social ou do nome fantasia, com resultados paginados.

## Requisitos

- Go 1.27.0 ou superior.
- MySQL com uma base carregada no formato compatível.
- Git para clonar o repositório.

O ambiente de referência utiliza MySQL 8.4.11 em Docker.

## Compatibilidade com o go-cnpj-loader

A API pode ser utilizada em conjunto com o
[go-cnpj-loader](https://github.com/danielsantello/go-cnpj-loader),
que prepara e carrega a base no formato esperado.

Os projetos são independentes. A API também pode consultar uma base
carregada por outras ferramentas, desde que ela siga o formato suportado,
incluindo tabelas, colunas e metadados esperados.

Na inicialização, a API verifica os metadados e a compatibilidade do
schema. Os índices necessários são criados por um comando da própria API.

## Instalação

Clone o repositório e instale as dependências:

```bash
git clone https://github.com/danielsantello/go-cnpj-api.git
cd go-cnpj-api
go mod download
```

### Configuração

Crie um arquivo `.env` na raiz do projeto:

```dotenv
CNPJ_API_HTTP_ADDRESS=:8080
CNPJ_API_SHUTDOWN_TIMEOUT=10s

CNPJ_API_MYSQL_HOST=127.0.0.1
CNPJ_API_MYSQL_PORT=3306
CNPJ_API_MYSQL_DATABASE=cnpj_2026_08_001
CNPJ_API_MYSQL_USER=seu_usuario
CNPJ_API_MYSQL_PASSWORD=sua_senha
CNPJ_API_MYSQL_CONNECT_TIMEOUT=5s
```

Ajuste a conexão para o seu ambiente. O nome do banco acima é um exemplo.

A aplicação lê variáveis de ambiente; o arquivo `.env` não é carregado
automaticamente. No Bash, carregue-o no terminal antes de executar:

```bash
set -a
source .env
set +a
```

### Criação dos índices

Com a base carregada e as variáveis configuradas, execute:

```bash
go run ./cmd/cnpj-api database create-indexes
```

O comando cria os índices ausentes. Execuções posteriores preservam
os índices já existentes.

### Execução

No mesmo terminal, inicie a API:

```bash
go run ./cmd/cnpj-api serve
```

Por padrão, ela estará disponível em `http://localhost:8080`.

Para encerrar, pressione `Ctrl+C`.

## Números observados

Referência: base da publicação 2026/08, utilizando MySQL 8.4.11
em container Docker.

| Indicador | Valor observado |
|---|---:|
| Estabelecimentos — registros estimados | 64,6 milhões |
| Empresas — registros estimados | 62,1 milhões |
| Índices no catálogo | 14 |
| Criação dos 11 índices iniciais | 11m02,542s |
| Índice de razão social | 5m09,641s |
| Índice de nome fantasia | 5m15s |
| Total acumulado de criação dos 14 índices | Aproximadamente 23m27s |
| Verificação com 12 índices já existentes | 0,596s |

O total de criação é uma soma aproximada de execuções distintas.
Esses tempos são referências observadas e variam conforme o ambiente
e o volume de dados.

Consulte o [levantamento completo dos índices](docs/indexes.md)
para os tempos por tabela e as condições de medição.

## Utilização

| Método | Endpoint | Finalidade |
|---|---|---|
| GET | `/health` | Verificar a disponibilidade da API e do MySQL |
| GET | `/v1/companies/{cnpj}` | Consultar dados pelo CNPJ |
| GET | `/v1/companies?name=SANTELLO` | Buscar por início da razão social |
| GET | `/v1/establishments?name=SANTELLO` | Buscar por início do nome fantasia |
| GET | `/v1/partners?name=SANTELLO` | Buscar sócios e empresas vinculadas |
| GET | `/v1/search?name=SANTELLO` | Buscar termos na razão social, nomes fantasia e sócios; requer Manticore |

Exemplo de busca:

```bash
curl -sS --get 'http://localhost:8080/v1/companies' \
  --data-urlencode 'name=SANTELLO' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

As buscas por nome utilizam prefixo: `SANTELLO` encontra nomes que
começam com esse texto.

Veja os [endpoints e exemplos de utilização](docs/endpoints.md).

## Testes

```bash
go test ./...
```

## Manticore Search — opcional

A API permite habilitar a conexão com o Manticore Search.
A busca ampla está disponível em `/v1/search`, sobre os documentos
já indexados. A carga completa e a troca de publicação ainda estão
em implementação.

Ambiente validado com Manticore Search 29.9.0 em Docker.

Para habilitar, acrescente às variáveis de ambiente:

```dotenv
CNPJ_API_MANTICORE_ENABLED=true
CNPJ_API_MANTICORE_HOST=127.0.0.1
CNPJ_API_MANTICORE_PORT=9306
CNPJ_API_MANTICORE_CONNECT_TIMEOUT=5s
```

Por padrão, a integração está desabilitada. As consultas e buscas
por prefixo continuam disponíveis usando apenas MySQL.

Quando habilitado, o Manticore é verificado na inicialização.
Se a conexão falhar, a API não inicia.

O endpoint `/health` também verifica o serviço habilitado:
`search.status` retorna `available` ou `unavailable`.
Se a verificação falhar, o endpoint retorna HTTP 503.
Quando desabilitado, retorna `disabled`.

### Carga inicial do índice de busca

Com MySQL e Manticore configurados e as variáveis carregadas, execute:

```bash
go run ./cmd/cnpj-api search index --limit=1000 --batch-size=100
```

O comando cria a tabela no Manticore, se necessário, e indexa até
mil empresas, em lotes de cem, seguindo a ordem do CNPJ básico.

Cada documento reúne razão social, nomes fantasia e nomes dos sócios
de uma empresa. O MySQL é utilizado somente para leitura.

| Opção | Padrão | Descrição |
|---|---:|---|
| `--limit` | 1000 | Quantidade máxima de empresas nesta execução; deve ser positiva |
| `--batch-size` | 100 | Empresas por lote; de 1 a 1000 |

Repetir a carga atualiza os documentos com os mesmos identificadores,
sem duplicá-los. O comando não remove documentos de publicações anteriores.

Nesta etapa, o índice é parcial. A busca HTTP está disponível sobre essa amostra. O procedimento de substituição completa da base ainda está em implementação.

Na amostra de 20/10/2026, com Manticore 29.9.0 e MySQL 8.4.11,
mil empresas foram processadas em 1,053s; a execução completa com
`time go run` levou 1,618s. Repetir a carga manteve mil documentos.

Esses tempos representam somente a amostra e não estimam a duração
da carga completa.

### Carga em uma tabela separada

Use `--table` para escolher o destino sem alterar a tabela consultada
pela API:

```bash
go run ./cmd/cnpj-api search index \
  --table=cnpj_companies_2026_08_test \
  --limit=1000 \
  --batch-size=100
```

O nome deve começar com `cnpj_companies_`, seguido de 1 a 48 caracteres:
letras minúsculas, números ou sublinhados.

### Carga completa

Use `--all` com uma tabela de destino explícita:

```bash
go run ./cmd/cnpj-api search index \
  --table=cnpj_companies_2026_08 \
  --all \
  --batch-size=100
```

`--all` e `--limit` não podem ser usados juntos. A carga completa lê
a base até o fim, em lotes.

Para uma nova publicação, utilize uma tabela nova. Reutilizar uma tabela
atualiza os documentos existentes, mas não remove empresas ausentes
da nova base.

Criar e carregar a tabela não a ativa para consultas. A API continua
consultando `cnpj_companies`; a troca da tabela ativa ainda será implementada.

### Retomada de carga

Se uma execução for interrompida, utilize o último `last_basic_cnpj`
registrado em uma mensagem de progresso concluído:

```bash
go run ./cmd/cnpj-api search index \
  --table=cnpj_companies_2026_08 \
  --all \
  --after=00001132 \
  --batch-size=100
```

O cursor acima é ilustrativo. Use o valor da sua execução, mantendo
a mesma tabela de destino e a mesma base MySQL.

A opção `--after` exige um CNPJ básico com oito dígitos e uma tabela
de destino explícita. Ela não verifica se todos os documentos anteriores
ao cursor estão presentes: não deve ser usada para iniciar uma tabela vazia.

O contador `processed` considera apenas os documentos da execução atual.
