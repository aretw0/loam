# ADR 00X: Análise Crítica e Adoção de CRDTs no Loam

**Data:** 19 de Fevereiro de 2026
**Status:** Proposta / Em Discussão

## Contexto e Análise Crítica (@/critical-analysis)

O **Loam** se posiciona como um motor de documentos transacional e reativo focado no paradigma *local-first*, armazenando dados em texto puro (Markdown, JSON, YAML) e usando Git como *backend* de versionamento e auditoria.

A premissa do projeto é fornecer controle total ao usuário sob a forma de arquivos. Entretanto, à medida que a necessidade de colaboração concorrente, peer-to-peer e sincronizações multiplataforma cresce (especialmente analisando o cenário de 2026 para aplicações *local-first*), nos deparamos com o limite do versionamento estritamente baseado em arquivos textuais tradicionais: **a resolução de conflitos estrutural**.

### Perspectivas

* **Perspectiva de Usuários em Potencial (PKM / Sync Apps):** A experiência *local-first* perfeita é aquela que parece "mágica". O usuário fica offline, faz edições extensas em seu aparelho móvel e no desktop simultaneamente. Ao conectar, ele espera que os dados se unam. O Git puro falha ao lidar com conflitos paralelos na mesma linha, especialmente em dados estruturados (esquemas JSON/YAML), exigindo resolução manual que corrompe a experiência do usuário não técnico.
* **Perspectiva Arquitetural e da Comunidade Go:** CRDTs (Conflict-free Replicated Data Types) tornaram-se o padrão-ouro na indústria para resolver esse exato gargalo. Bibliotecas amadureceram (Loro, Automerge, Yjs) e a comunidade Golang tem demandado integrações robustas para motores reativos como o Loam. O Loam hoje entrega a reatividade e a transação local; adicionar garantias de mesclagem (merge) sem conflitos é o próximo passo para alcançar relevância de alto nível e concorrência distribuída descentralizada.

## Respondendo aos Questionamentos Iniciais

### 1. O Loam deveria se aproximar das abordagens de CRDTs?

**Sim, mas como um complemento, não como um substituto.**
O valor central do Loam é manter os dados em texto puro com trilha de autoria clara via Git. Se utilizarmos *blobs* binários inescrutáveis (comuns em motores CRDT estruturais), quebraremos a promessa de transparência. O Loam deve adotar CRDTs especificamente para a **resolução automática de merges** concorrentes, mantendo o "snapshot" final perfeitamente legível e acessível em formato de texto.

### 2. O que podemos aprender com as aplicações atuais (2026)?

A análise da evolução de ferramentas (como Loro, Automerge, Yjs e Dolt) revela que:

* **Sinergia evita atrito:** Criar soluções de versionamento do zero é menos adotado do que estender ferramentas sedimentadas, como o Git.
* **Intenção sobre histórico técnico:** O foco de um bom CRDT textual é preservar a intenção do autor sem onerar o histórico com entrelaçamentos (*interleavings*) confusos.
* **Separação de estados:** Sistemas de sucesso separam claramente a camada "estado atual" (o Markdown legível) do "oplog/operações" (o estado isolado do CRDT).
* Portanto, o modelo vencedor não infla o repositório original. O Git mantém os *checkpoints* textuais, enquanto a lógica de CRDT atua invisível para garantir a consolidação do conflito.

### 3. Fazer à parte ou integrar no ecossistema do Loam?

Desenvolver como um projeto isolado fragmentaria esforços. A arquitetura do Loam foi pensada em *Adapters*, o que facilita essa incorporação. Há dois caminhos fundamentais para tratar esses conflitos:

#### A. O Adapter Customizado (Reconciliação na Aplicação)

* **Como funciona:** O Loam gerencia o texto e o estado do CRDT utilizando um adaptador independente (ex: `pkg/adapters/crdt`). O *log* do CRDT reside numa pasta oculta. Durante a sincronização, o código do Loam intercepta os arquivos e efetua o merge estrutural antes do Git atuar.
* **Prós:** Traz controle total para o código Go, simplificando depuração e removendo a dependência de chamadas à CLI do Git do sistema em que opera.
* **Contras:** Cria dependência crônica *(lock-in)*. Se o usuário rodar um `git pull` genérico no terminal, alheio aos comandos do Loam, o Git fará a consolidação padrão de texto e alertará conflitos estruturais (`<<<<<<<`).

#### B. O Custom Merge Driver (Educar o Git)

* **Como funciona:** Injetamos um *Merge Driver* customizado no arquivo `.gitattributes` para as extensões suportadas (`.json`, `.md`). Quando o Git processa um *commit* concorrente, ele delega a resolução (*3-way merge*) diretamente para o binário do Loam, que compõe a versão estabilizada.
* **Prós:** Funciona de forma transparente. O usuário ou outras aplicações (ex: GitHub Desktop) seguem com os seus fluxos do dia a dia nativamente e o repositório é fundido de maneira limpa. Transforma o Git numa plataforma real para a nossa extensão.
* **Contras:** Fricção de configuração inicial (exige rotinas de *bootstrap* para injetar comandos no `.git/config` da máquina) e a complexidade técnica extra de passar contexto externo (os arquivos do CRDT) caso demandados junto com a modificação nativa processada pelo Git.

#### Convenção Recomendada

A recomendação técnica pautada nas referências atuais do mercado guia as prioridades para a integração via **Custom Merge Driver (Abordagem B)** coordenada em segundo plano pelo `Adapter fs` na inicialização do repositório. Isso alinha perfeitamente com a tese de "magia invisível" ao usuário alvo.

## Decisões & Próximos Passos (Proposta)

1. **Documentação de Arquitetura:** Fundação das ADRs sob a pasta `docs/architecture` formalizada por meio deste artefato.
2. **Experimentação Prática (PoC):** Iniciar as provas de conceito de um sub-comando (*ex: `loam merge-driver`*) integrando de modo contíguo uma biblioteca CRDT leve para a mediação perfeitamente em plano de fundo.
3. **Foco Tático:** Validar como se daria a injeção automatizada via `.git/config` no *Adapter* atual (`fs`) observando a viabilidade de adotar resolução assíncrona textualmente sem destruir os próprios comandos do sistema.
