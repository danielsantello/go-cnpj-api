# Operação das buscas

## Contratos e escolha do mecanismo

| Endpoint | Unidade do resultado | Campo pesquisado no Manticore |
|---|---|---|
| `/v1/companies` | Empresa, com um CNPJ representante | `corporate_name` |
| `/v1/establishments` | Estabelecimento encontrado | `trade_name` |
| `/v1/partners` | Vínculo entre sócio e empresa | `partner_name` |
| `/v1/search` | Empresa com campos de correspondência | Razão social, nomes fantasia e sócios agregados |

Os três endpoints específicos usam termos no Manticore quando habilitado
e disponível. Todos os termos são obrigatórios no campo correspondente,
em qualquer ordem e posição. Se desabilitado ou indisponível, a consulta
inteira é refeita pelo início do nome no MySQL. Não há combinação dos
itens ou dos totais dos dois mecanismos.

Falhas de conexão, timeout de comunicação e tabela ausente permitem
fallback. Resultado vazio, validação, `SEARCH_TOO_BROAD`, erro SQL
inesperado, falha de enriquecimento e cancelamento não permitem fallback.
Uma tabela parcial existente é considerada disponível: ela pode retornar
vazio sem que o MySQL seja consultado como alternativa.

Cada endpoint conta, aplica o limite de 1000 e pagina sua própria entidade.
Em sócios, `document` é uma condição AND por igualdade, antes dessas etapas;
a máscara é literal. Homônimos e vínculos repetidos não são fundidos.
O filtro de documento não está disponível na busca ampla `/v1/search`,
que mantém seu contrato separado e exige Manticore.

Município, UF e situação cadastral correspondem ao CNPJ completo retornado.
Com Manticore, somente os CNPJs da página são enriquecidos em um lote no
MySQL, preservando ordem e paginação. Páginas vazias não são enriquecidas.

A ordem no Manticore é pelo ID do documento. No MySQL, é pelo nome e
identificador correspondente, conforme descrito em [endpoints](endpoints.md).
Uma mudança de mecanismo pode mudar ordem e resultados entre páginas.

## Configuração e disponibilidade

```dotenv
CNPJ_API_MANTICORE_ENABLED=true
CNPJ_API_MANTICORE_HOST=127.0.0.1
CNPJ_API_MANTICORE_PORT=9306
CNPJ_API_MANTICORE_CONNECT_TIMEOUT=5s
CNPJ_API_MANTICORE_TABLE=cnpj_companies_2026_08
CNPJ_API_MANTICORE_ESTABLISHMENTS_TABLE=cnpj_establishments_2026_08_v2
CNPJ_API_MANTICORE_PARTNERS_TABLE=cnpj_partners_2026_08_v2
```

Configure as tabelas acima somente após concluir suas cargas. A tabela
por empresa atende razão social e busca ampla. O índice agregado anterior
continua compatível para leitura, inclusive para razão social; não substitui
os índices de estabelecimentos ou de vínculos de sócios.

O servidor HTTP mantém o pool e inicia mesmo se o ping do Manticore falhar.
Novas requisições podem reconectar sem reiniciar a API. No servidor HTTP,
o timeout configurado também limita cada leitura/escrita de socket para
que uma conexão sem resposta possa permitir fallback. Não é um limite
único de duração da requisição: podem existir várias operações e tentativas.
Os comandos de carga e remoção continuam exigindo conexão disponível e
não recebem esse limite adicional de leitura/escrita.

`/health` verifica conexão com os serviços habilitados, retornando HTTP 503
se algum falhar, mesmo quando as buscas específicas continuam pelo MySQL.
O health não comprova existência, completude ou publicação das tabelas.

## Carga simples, sempre do início

Cada comando de carga apaga e recria a tabela de destino antes de ler a
origem. A carga usa lotes para limitar memória e atribui IDs sequenciais,
começando em 1. Os CNPJs permanecem strings, sem conversão numérica.

Se a carga falhar ou for interrompida, corrija a causa e execute o mesmo
comando novamente. A tabela parcial será eliminada e a leitura começará
do zero. Não há retomada, opções de checkpoint, verificação de proprietário
de cada ID ou marcador de versão por documento. Os cursores entre lotes
são somente internos à execução.

A aplicação do código não altera as tabelas existentes. Os índices já
carregados continuam compatíveis para leitura, inclusive os que possuem
colunas extras de uma versão anterior. Somente a execução de um comando
de carga recria seu destino; não o execute para apenas ativar uma carga pronta.

A tabela por empresa possui um documento por CNPJ básico e agrega nomes
fantasia e de sócios para a busca ampla. A tabela por estabelecimento
possui um documento por CNPJ completo com nome fantasia não nulo.
A tabela por sócio possui um documento por vínculo, exige empresa e
estabelecimento associados e preserva documentos nulos e vazios.
Vínculos com nome e documento iguais continuam separados.

Os comandos exigem `--table` e recusam qualquer uma das três tabelas
configuradas como ativas. Use o mesmo ambiente do servidor para que essa
proteção corresponda às tabelas em uso. Ela não identifica outros servidores
com configuração diferente nem outras cargas em andamento. Execute uma
única carga por tabela e mantenha a publicação MySQL estável durante a leitura.
Dados duplicados que inviabilizem identificar uma empresa ou estabelecimento
interrompem a carga; os registros da origem não são corrigidos.

Carregue as variáveis no terminal:

```bash
clear
cd /home/ops/projects/go-cnpj-api
set -a
source .env
set +a
```

### Cargas completas

Use tabelas de destino inativas. Se já existirem, seu conteúdo será apagado:

```bash
clear
go run ./cmd/cnpj-api search index-establishments \
  --table=cnpj_establishments_2026_08_v2 \
  --all --batch-size=1000
```

```bash
clear
go run ./cmd/cnpj-api search index-partners \
  --table=cnpj_partners_2026_08_v2 \
  --all --batch-size=1000
```

O índice completo de empresas já existente pode continuar atendendo razão
social e `/v1/search`. Para carregar uma nova tabela por empresa:

```bash
clear
go run ./cmd/cnpj-api search index \
  --table=cnpj_companies_2026_08_v2 \
  --all --batch-size=1000
```

Todos os comandos aceitam `--batch-size` de 1 a 1000, com padrão 100.
`--all` lê toda a origem elegível. Para uma amostra, use `--limit=1000`
no lugar de `--all`, com um destino diferente, por exemplo uma tabela
com sufixo `_sample`. Sem `--all`, o limite padrão é 1000.
`--all` e `--limit` não podem ser combinados.

As opções `--after` e `--after-id` foram removidas. Os logs mostram o
número de documentos processados e o tempo decorrido. Não há ativação
automática ao concluir a carga.

### Ativação e remoção

Depois de concluir e validar contagens, vínculos, CNPJs e consultas,
altere as variáveis das tabelas correspondentes e reinicie a API.
Mantenha os índices anteriores até concluir a validação; para voltar,
restaure a configuração anterior e reinicie.

O nome da tabela não comprova que ela está completa ou que corresponde
ao MySQL. As três tabelas ativas devem representar a mesma publicação.

O comando de remoção aceita as três famílias e protege todas as tabelas
ativas da configuração carregada:

```bash
clear
go run ./cmd/cnpj-api search drop \
  --table=cnpj_partners_2026_08_test
```

A remoção é permanente. Use o mesmo ambiente do servidor. A proteção não
detecta outra instância com configuração diferente nem uma carga em andamento.

## Medições e validações observadas

Em 23/10/2026, publicação 2026/08, MySQL 8.4.11 e Manticore 29.9.0 locais:

| Amostra | Registros | Rotina | Processo completo |
|---|---:|---:|---:|
| Estabelecimentos | 1000 | 144,968 ms | 600 ms |
| Vínculos de sócios | 1000 | 234,449 ms | 708 ms |

Foram conferidos os IDs sequenciais e dez registros de cada amostra
contra os detalhes da API; os dez vínculos de sócios envolveram quatro CNPJs.
A otimização da leitura dos sócios está em [índices](indexes.md).
São amostras; não estimam duração de carga completa nem desempenho HTTP.

A validação HTTP confirmou prefixo no MySQL com Manticore desabilitado,
início da API com Manticore inacessível, reconexão e queda posterior sem
reiniciar a API, filtro de documento, campos e páginas independentes.
Na base completa por empresa, `SANTELLO` retornou 209 empresas por termos.
Nas amostras, `GERAL` retornou um estabelecimento e `OLIVEIRA`, 51 vínculos.
Resultados vazios nesses índices não acionaram fallback, mesmo havendo
resultados por prefixo no MySQL.

Os testes de integração cobrem restrição de campo, 1000/1001 resultados,
vínculos repetidos, homônimos, documentos literais/nulos/vazios, CNPJs
alfanuméricos e recriação do destino, removendo registros anteriores e
reiniciando os IDs. Para executá-los,
use uma conexão de teste com permissão para criar/remover tabelas temporárias
de teste no Manticore e tabelas temporárias de sessão no MySQL:

```bash
clear
CNPJ_API_RUN_INTEGRATION_TESTS=1 go test \
  ./internal/storage/manticore ./internal/storage/mysql/v1 -count=1 -v
```

### Carga completa de estabelecimentos concluída

Em 23/10/2026, `cnpj_establishments_2026_08_v2` terminou com 26.523.417
estabelecimentos elegíveis e último CNPJ `99017782000139`.

| Medição | Tempo |
|---|---:|
| Rotina, conforme log | 53m32,249s |
| Processo completo, conforme `time` | 59m27,475s |

Essa tabela foi carregada antes da simplificação, com uma coluna adicional
de versão. Ela continua compatível com os leitores atuais e não precisa
ser recarregada por causa da mudança do comando. O término da carga ainda
precisa ser acompanhado da conferência das contagens e dados antes da ativação.

## Validação final com as tabelas completas — 24/10/2026

Tabelas ativadas:

- Empresas: `cnpj_companies_2026_08`.
- Estabelecimentos: `cnpj_establishments_2026_08_v2`.
- Vínculos de sócios: `cnpj_partners_2026_08_v2`.

A carga de sócios concluiu com 28.146.721 vínculos:
51m29,699s na rotina e 57m15,127s no processo completo.
O total foi conferido no Manticore e seis vínculos do início e do fim
foram comparados com os detalhes de três CNPJs.

A carga de estabelecimentos concluiu com 26.523.417 documentos.
O total foi conferido no Manticore e seis estabelecimentos do início
e do fim foram comparados com os detalhes da API. A contagem independente
dos registros elegíveis no MySQL não foi concluída: a consulta perdeu
a conexão do cliente após 6000 segundos e foi posteriormente cancelada
no servidor. Não há confirmação independente desse total na origem.

Na consulta `SANTELLO`, foram conferidos contagem, ordem e duas páginas
de cada endpoint contra o Manticore:

| Busca | Resultados |
|---|---:|
| Razão social | 209 |
| Nome fantasia | 29 |
| Vínculos de sócios | 219 |
| Sócios com nome AND documento da amostra | 12 |

Município, UF e situação cadastral foram comparados com os detalhes
do estabelecimento correspondente em 33 CNPJs distintos.

### Tempos HTTP observados

Uma requisição de aquecimento e cinco medições sequenciais por combinação,
na página 1. Todas retornaram HTTP 200. Tempos em milissegundos, medidos
por um cliente Python, incluindo leitura e decodificação do JSON.

| Busca | Tamanho solicitado | Itens retornados | Mediana | Mínimo | Máximo |
|---|---:|---:|---:|---:|---:|
| Razão social | 20 | 20 | 4,576 | 4,552 | 4,814 |
| Razão social | 100 | 100 | 7,828 | 7,715 | 8,750 |
| Nome fantasia | 20 | 20 | 4,457 | 4,072 | 5,312 |
| Nome fantasia | 100 | 29 | 4,170 | 4,167 | 4,434 |
| Sócios | 20 | 20 | 3,996 | 3,945 | 4,230 |
| Sócios | 100 | 100 | 6,473 | 5,905 | 7,953 |
| Busca ampla | 20 | 20 | 6,966 | 6,722 | 7,231 |
| Busca ampla | 100 | 100 | 10,353 | 9,263 | 46,308 |

A busca ampla retornou 321 empresas. Os universos por termos diferem
das antigas buscas por prefixo; os tempos não constituem uma comparação
direta de desempenho entre os mecanismos. A amostra não mede concorrência,
carga sustentada ou latência de cauda.
