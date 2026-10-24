# go-cnpj-api

API em Go para consulta de dados públicos de CNPJ da Receita Federal.

Permite consultar empresas pelo CNPJ e pesquisar razão social, nome
fantasia e sócios. As três buscas usam termos com Manticore disponível
e prefixo no MySQL quando desabilitado ou indisponível. Uma busca ampla
separada reúne termos dos três campos por empresa.

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

Consulte o [formato da base e dicionário de dados](https://github.com/danielsantello/go-cnpj-loader/blob/main/docs/referencia/formato-da-base.md)
para os campos, tipos, relacionamentos e regras de importação do formato 1.

Na inicialização, a API verifica os metadados e a compatibilidade do
schema. Os índices necessários são criados por um comando da própria API.

## Instalação

Clone o repositório e instale as dependências:

```bash
clear
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
clear
set -a
source .env
set +a
```

### Criação dos índices

Com a base carregada e as variáveis configuradas, execute:

```bash
clear
go run ./cmd/cnpj-api database create-indexes
```

O comando cria os índices ausentes. Execuções posteriores preservam
os índices já existentes.

### Execução

No mesmo terminal, inicie a API:

```bash
clear
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
| Empresas — contagem exata após correção de uma duplicata | 69.523.303 |
| Documentos no índice completo do Manticore | 69.523.303 |
| Índices no catálogo | 15 |
| Índice de seleção do estabelecimento representante | 30m16,418s |
| Criação dos 11 índices iniciais | 11m02,542s |
| Índice de razão social | 5m09,641s |
| Índice de nome fantasia | 5m15s |
| Total acumulado de criação dos 14 índices anteriores | Aproximadamente 23m27s |
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
| GET | `/v1/companies?name=SANTELLO` | Buscar razão social por termos ou prefixo |
| GET | `/v1/establishments?name=SANTELLO` | Buscar nome fantasia por termos ou prefixo |
| GET | `/v1/partners?name=SANTELLO` | Buscar sócios e empresas vinculadas |
| GET | `/v1/search?name=SANTELLO` | Buscar termos na razão social, nomes fantasia e sócios; requer Manticore |

Exemplo de busca:

```bash
clear
curl -sS --get 'http://localhost:8080/v1/companies' \
  --data-urlencode 'name=SANTELLO' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=5'
```

As buscas específicas selecionam o mecanismo automaticamente e mantêm
contagens e paginações independentes. O parâmetro opcional `document`
restringe somente a busca de sócios, por igualdade e como condição AND.
A busca ampla exige Manticore. Todas permitem até 1000 resultados;
acima disso, retornam HTTP 422, solicitando refinar a busca.

Veja os [endpoints e exemplos de utilização](docs/endpoints.md).

## Testes

```bash
clear
go test ./...
```

## Manticore Search — opcional

Ambiente validado com Manticore Search 29.9.0 em Docker.

```dotenv
CNPJ_API_MANTICORE_ENABLED=true
CNPJ_API_MANTICORE_HOST=127.0.0.1
CNPJ_API_MANTICORE_PORT=9306
CNPJ_API_MANTICORE_TABLE=cnpj_companies_2026_08
CNPJ_API_MANTICORE_ESTABLISHMENTS_TABLE=cnpj_establishments_2026_08_v2
CNPJ_API_MANTICORE_PARTNERS_TABLE=cnpj_partners_2026_08_v2
CNPJ_API_MANTICORE_CONNECT_TIMEOUT=5s
```

Configure as tabelas somente depois de concluir e validar suas cargas.
O índice por empresa atende razão social e `/v1/search`; os demais atendem
estabelecimentos e vínculos de sócios, respectivamente. Os padrões são
`cnpj_companies`, `cnpj_establishments` e `cnpj_partners`.

Por padrão, Manticore está desabilitado. Quando habilitado e inacessível,
o servidor HTTP inicia e as três buscas específicas podem usar MySQL.
O pool permite reconexão sem reiniciar a API. O timeout configurado limita
conexão e operações de leitura/escrita do Manticore no servidor HTTP;
não é um prazo único da requisição, que pode executar várias operações.
Os comandos de carga continuam exigindo uma conexão disponível.

O health verifica a conexão: retorna HTTP 503 se o Manticore habilitado
estiver indisponível, mesmo quando as buscas específicas usam fallback.
Ele não comprova que as tabelas existem ou que estão completas.

Consulte [carga, ativação e remoção dos índices](docs/search.md).
Os índices existentes continuam compatíveis para leitura. Os comandos de
carga exigem um destino explícito e inativo, que é apagado e recriado
em cada execução. Se houver falha, execute novamente desde o início;
não há retomada ou checkpoint persistente.

### Histórico da carga completa anterior — publicação 2026/08

Ambiente: MySQL 8.4.11 e Manticore Search 29.9.0 em Docker,
em máquina com aproximadamente 7,7 GiB de RAM.

A carga foi concluída com retomadas e lotes de 1000 empresas.
As contagens finais foram iguais: 69.523.303 empresas no MySQL
e 69.523.303 documentos no Manticore.

Na última execução, foram processadas 61.455.303 empresas:

| Medição da última execução | Tempo |
|---|---:|
| Rotina de indexação, conforme log | 1h44m12,977s |
| Processo completo, conforme `time` | 1h53m09,754s |

Esses tempos não representam a carga inteira: 8.068.000 empresas
já haviam sido processadas em execuções anteriores.

Durante a carga, o acesso à tabela de empresas foi ajustado para usar
o índice de CNPJ básico por intervalo. Também foi removido um registro
duplicado com nome vazio, preservando o registro preenchido.

Os números são referências observadas, não garantias de desempenho.

Esse histórico é do comando anterior. O comando atual sempre recria seu
destino e começa do zero. O índice já carregado continua compatível para leitura.
