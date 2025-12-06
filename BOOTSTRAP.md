# BOOTSTRAP

Este arquivo serve para fazer o bootstrap deste projeto chamado **Loam** (nome provisório).

## Objetivos

O objetivo principal é fornecer uma **camada de persistência transacional** para arquivos Markdown com Frontmatter, atuando como um "driver de banco de dados" para sistemas de arquivos locais.

Metas específicas:

1. **Centralizar a Lógica de I/O:** Abstrair operações de leitura, escrita e *parsing* para evitar duplicação de código em diferentes ferramentas.
2. **Garantir Integridade (ACID-ish):** Prevenir condições de corrida (*race conditions*) quando múltiplos processos tentam editar o mesmo cofre de notas simultaneamente.
3. **Versionamento como Log:** Utilizar o Git (se disponível) como um *Write-Ahead Log* transparente, garantindo histórico e reversibilidade atômica.
4. **Agnosticismo:** Funcionar independentemente do editor utilizado (Obsidian, VS Code, Vim), focando apenas na estrutura dos dados (Markdown + YAML).
5. **Portabilidade:** Ser distribuído como uma biblioteca Go (embeddable) e/ou um binário *standalone*, facilitando a criação de ecossistemas em diversas linguagens.

## Visão Geral

O **Loam** é um *engine* (motor) escrito em Go que trata uma pasta de arquivos Markdown como um banco de dados NoSQL. Ele oferece uma API para operações de CRUD, mas com superpoderes de gerenciamento de concorrência e versionamento.

Diferente de um banco de dados tradicional que esconde os dados em arquivos binários, o Loam orgulha-se de manter os dados em **texto plano legível por humanos**, enriquecendo a experiência com:

- **Transações em Lote:** Capacidade de agrupar múltiplas alterações em um único commit.
- **Commit Gates:** Mecanismos opcionais de espera/confirmação antes de efetivar alterações no Git (permitindo revisão de *diffs*).
- **Indexação Leve:** Cacheamento de metadados para buscas rápidas sem ler todo o disco repetidamente.

### Público-alvo

1. **Desenvolvedores de Ferramentas (Toolmakers):** Pessoas criando bots, CLIs ou automações para gestão de conhecimento pessoal (PKM) ou empresarial.
2. **Engenheiros de Dados Pessoais:** Usuários avançados que desejam pipelines de ETL (Extração, Transformação e Carga) para suas notas (ex: Financeiro -> Markdown).
3. **Entusiastas de "Local-First":** Quem busca soberania sobre seus dados, recusando bancos de dados proprietários ou formatos binários fechados.

## Abordagem

Seguiremos uma abordagem incremental para evitar *over-engineering*:
**Spike -> SDD Simplificado -> (Dev Container) -> BDD/TDD**

### 1. Fase 0: Spike (Prova de Conceito)

**Objetivo:** Validar a viabilidade técnica da stack Go + Git + Concorrência.
*Por que?* Precisamos provar que conseguimos gerenciar *locks* de arquivos e invocar comandos Git de forma confiável e performática antes de construir abstrações complexas. O Spike deve resultar em um pequeno programa que escreve em 100 arquivos simultaneamente e garante que o `git status` final esteja limpo e consistente.

**Critérios de Sucesso Adicionais:**

- **Teste de "Dirty State":** Verificar comportamento se o diretório já estiver sujo antes da execução.
- **Teste de "File Watching":** Garantir que a escrita atômica não dispare eventos excessivos em watchers externos.

### 2. SDD - Specification Driven Development (Simplificado)

Evitar a "burocracia de documentação". Focaremos em poucos arquivos vivos em `/docs`:

- `PRODUCT.md`: Visão, Personas e User Stories (O "Porquê" e "Para Quem").
- `TECHNICAL.md`: Arquitetura, Decisões de Design (Go, gRPC/Embed, File Locking) e Stack.
- `PLANNING.md`: Roadmap, Backlog e Tarefas imediatas.

*Na raiz:*

- `README.md`: Guia de início rápido e exemplos de uso da lib.

### 3. Dev Container & Ambiente

Definir um ambiente reprodutível (Dev Container) apenas após validar o Spike, para garantir que outros devs possam colaborar sem "works on my machine". O ambiente deve incluir Go, Git e ferramentas de linting.

### 4. Qualidade (BDD/TDD)

- **BDD:** Testes de aceitação para garantir que o fluxo "Transação Iniciada -> Escrita -> Commit -> Confirmação" funcione como esperado pelo usuário final.
- **TDD:** Testes unitários rigorosos para o *Core*, especialmente para o parser de Frontmatter e o gerenciador de filas/locks, utilizando diretórios temporários (`t.TempDir`) para não poluir o repositório do projeto.

## Premissas e Restrições Técnicas

1. **Latência "Humana":** Aceitamos que o Git será o gargalo. A performance não precisa competir com bancos SQL, mas deve ser aceitável para interações humanas (segundos, não milissegundos).
2. **Acesso Exclusivo (Single-Tenant):** O Loam assume que é o único processo escrevendo ativamente na pasta durante uma transação. Não lidaremos com *locking* complexo entre múltiplos usuários simultâneos no SO.
3. **Forward Only (Sem Rollback de Commit):** O Loam nunca reverte um commit já realizado no Git ("git revert"). Se uma transação falhar *antes* do commit, descartamos as mudanças nos arquivos. Se falhar *durante/após*, o commit persiste. A consistência é garantida apenas *checkpoints* (commits).

### Próximos Passos Imediatos

1. [ ] Criar a estrutura de pastas inicial e o `go.mod`.
2. [ ] Criar `docs/PLANNING.md` com o backlog inicial focado no Kernel.
3. [ ] Executar o **Spike**: Criar um script Go que inicializa um repo git temporário e realiza commits em paralelo.
