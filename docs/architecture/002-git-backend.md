# ADR 002: Git como Motor de Armazenamento Central

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

Para o controle de versões, metadados locais de sistemas integrados ou diários pessoais (PKM), o histórico é tão valioso quanto o próprio dado atual.
Sistemas que provêm auditoria ou gestão temporal de conteúdo costumam embutir bancos de dados robustos (SQLite, RocksDB) ou dependem intensivamente de APIs de terceiros. Gostaríamos de fornecer um motor local-first com histórico auditável mas sem aprisionar os dados em estruturas exclusivas.

## Decisão

O Loam adota o **Git como Backend de Armazenamento e Histórico**. Em sincronia com o *File System Adapter*, o Git funciona não apenas como Backup Remoto, mas fundamentalmente como o *Write-Ahead Log* transacional para os recursos persistidos.
Além disso, a operação é projetada para ser tolerante a instâncias onde o Git falhe ou não exista: o `Smart Gitless Mode` permite que o sistema detete sua ausência, ativando armazenamento apenas em disco simples (`.loam/`).

## Consequências

* **Pró (Liberdade de Dados):** "Fidelidade". O usuário é de fato dono do conteúdo armazenado, podendo rodar comandos padrão do Git (como *log*, *blame*, *revert*) sem necessitar usar as APIs do Loam.
* **Pró (Auditoria Gratuita):** Trilha de modificações com assinaturas textuais prontas, permitindo identificar com exatidão autoria no tempo.
* **Contra (Overhead e Sincronização):** Gerenciar conflitos de texto massivos (exigindo estratégias como o futuro *Custom Merge Driver*). Adicionalmente, chamadas *exec* sequenciais na linha de comando do Git podem apresentar gargalos computacionais se o volume de transações por segundo (TPS) for altíssimo.
* **Contra (Event Storm):** Necessidade iminente de construir um "Git Lock Monitor" interno para ignorar oscilações no Filesystem que ocorrem em massa enquanto o `.git/index.lock` estiver aberto durante rebases ou checkouts profundos.
