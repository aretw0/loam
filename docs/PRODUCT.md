# Visão do Produto

**Loam** é um "driver de banco de dados" para sistemas de arquivos locais.
O objetivo principal é fornecer uma **camada de persistência transacional** para arquivos Markdown com Frontmatter.

## Objetivos

1. **Centralizar a Lógica de I/O:** Abstrair operações de leitura, escrita e *parsing* para evitar duplicação de código em diferentes ferramentas.
2. **Garantir Integridade (ACID-ish):** Prevenir condições de corrida quando múltiplos processos tentam editar o mesmo cofre de notas simultaneamente.
3. **Versionamento como Log:** Utilizar o Git como um *Write-Ahead Log* transparente, garantindo histórico e reversibilidade atômica.
4. **Agnosticismo:** Funcionar independentemente do editor (Obsidian, VS Code, Vim), focando apenas na estrutura dos dados (Markdown + YAML).
5. **Portabilidade:** Ser distribuído como uma biblioteca Go e/ou um binário *standalone*.

## Personas (Público-alvo)

1. **Toolmakers:** Desenvolvedores criando bots, CLIs ou automações para PKM (Personal Knowledge Management).
2. **Engenheiros de Dados Pessoais:** Usuários avançados que desejam pipelines de ETL para suas notas.
3. **Entusiastas de "Local-First":** Quem busca soberania sobre seus dados, recusando bancos proprietários.

## User Stories

- "Como desenvolvedor, quero garantir que minhas automações não corrompam o repositório Git (concorrência interna)."
- "Como usuário, quero desfazer um script mal sucedido usando `git revert` sem perder o estado consistente do cofre."
- "Como ferramenta externa, quero ler o frontmatter de 1000 notas rapidamente."

## Filosofia de Design

### Commits Semânticos Abstraídos

O Loam deve tratar o histórico do Git como um log estruturado, não apenas texto livre.

- **Abstração:** O usuário não escreve mensagens de commit livres (`-m "fixed stuff"`). Ele informa a *intenção* (Feat, Chore, Fix) e um título.
- **Padronização:** O Loam gera a mensagem final seguindo Conventional Commits, garantindo que feeds de changelog sejam gerados automaticamente.
- **Assinatura:** Commits gerados via CLI/Driver devem conter um footer `Footer: Powered by Loam`, indicando a origem da mudança.
