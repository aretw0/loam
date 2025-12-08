# Visão do Produto

**Loam** é um "driver de banco de dados" para conteúdo e metadados.
O objetivo principal é fornecer uma **camada de persistência transacional** agnóstica. A implementação de referência (FS Adapter) suporta nativamente **Markdown, JSON, YAML e CSV**, tratando arquivos como documentos estruturados.

## Por que Loam?

Para desenvolvedores acostumados com bancos de dados tradicionais, lidar com arquivos locais pode parecer arcaico e inseguro. O Loam traz a robustez de um banco de dados para o mundo dos arquivos de texto.

- **SQLite para Documentos:** Assim como o SQLite é o padrão para dados relacionais locais, o Loam quer ser o padrão para conteúdo não-estruturado (Markdown) e estruturado (CSV, JSON). Oferece suporte nativo a **Multi-Documentos** (coleções em um único arquivo) e um padrão **Active Record** para DX superior.
- **Automação Segura:** Seus scripts Python/Bash que editam notes podem corromper o repositório se rodarem concorrentemente. O Loam implementa *file locking* e transações para evitar isso.
- **CI/CD de Conteúdo:** Garanta integridade de dados (tags, datas, links) no momento da escrita, validando *frontmatter* antes do commit.

## Objetivos

1. **Centralizar a Persistência:** Abstrair operações de armazenamento e serialização para evitar duplicação de regras em diferentes ferramentas.
2. **Garantir Integridade (ACID-ish):** Prevenir condições de corrida quando múltiplos processos tentam editar o mesmo repositório simultaneamente.
3. **Histórico Auditável:** Manter um log de alterações transparente e reversível (implementado via Git no adapter padrão).
4. **Estrutura Universal:** Focar na estrutura "Conteúdo + Metadados", independente do formato de serialização final (Markdown, JSON, SQL).
5. **Portabilidade:** Ser distribuído como uma biblioteca Go e/ou um binário *standalone*.

## Personas (Público-alvo)

1. **Toolmakers:** Desenvolvedores criando bots, CLIs ou automações.
2. **Engenheiros de Dados:** Usuários avançados que desejam pipelines de ETL locais.
3. **DevOps & SREs:** Gerenciamento seguro de configurações (GitOps) e dotfiles.
4. **Entusiastas de "Local-First":** Quem busca soberania sobre seus dados.

## User Stories

- "Como desenvolvedor, quero garantir que minhas automações não corrompam o repositório Git (concorrência interna)."
- "Como usuário, quero desfazer um script mal sucedido usando `git revert` sem perder o estado consistente do cofre."
- "Como ferramenta externa, quero ler o frontmatter de 1000 notas rapidamente."

## Filosofia de Design

### Rastreabilidade Semântica

O Loam trata o histórico de mudanças como um log estruturado, não apenas texto livre.

- **Intenção sobre Implementação:** O usuário informa a *intenção* (Feat, Chore, Fix) e a razão da mudança.
- **Adaptação:** No adapter FS, isso se traduz em **Commits Semânticos** (Conventional Commits). Em um adapter SQL, poderia ser uma tabela de auditoria.
- **Assinatura:** Mudanças geradas via automação devem indicar sua origem (ex: `Powered by Loam`).
