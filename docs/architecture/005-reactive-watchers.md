# ADR 005: Arquitetura de Watchers Reativos

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

Um motor de documentos moderno não atua apenas como repositório morto de queries. O cenário de "Live Apps" exige que a interface de usuário reaja de forma imediata quando um conteúdo sobre a gestão do Loam for modificado (ex: o usuário altera um arquivo pelo terminal e a aplicação exibindo-o web é notificada e mostra o novo dado ao vivo).

## Decisão

Adotamos uma arquitetura de *Event Loop* protegido para os Watchers, implementada majoritariamente pelo pacote nativo acoplado a abstrações robustas:

1. **Monitoramento Recursivo em FS**: O `Watcher` base (`fsnotify`) engatilha callbacks monitorando toda a árvore sob gestão, integrado como uma `Source` nativa para o Lifecycle Control Plane.
2. **Git Awareness (Evitando Event Storms)**: Se o diretório base for um repositório Git, a Source monitora agressivamente as chamadas ao `index.lock`. Enquanto o Git executa transações densas em massa, a Source aplica **Inibição Local**, retendo as notificações nativamente (`Pause`).
3. **Self-Healing & Checksum Filters**: Quando a própria aplicação escreve os arquivos, isso é monitorado pelo OS e resultaria em eventos recursivos (`Eco`). Através da junção de uma Janela Temporal (< 2s) e Hashing (SHA256) dos *writes* autorais, o Watcher amortece os eventos em *Self-Writes*, descartando-os sem engatilhar o ciclo final do broker para os clientes.

## Consequências

* **Pró (Aplicações Vivas):** O sistema tem latência sub-segundo informando os clientes sobre modificações ambientais ou de colaboração.
* **Pró (Proteção Termodinâmica):** Os "debouncers" de 50ms associados à agregação (`MergedEvent`) processada e gerenciada nativamente pelo middleware do **Lifecycle Control Plane** protegem a performance da RAM da aplicação de picos excessivos do sistema operacional, mantendo a responsividade do host sem perder reatividade por arquivo.
* **Contra (Limitações do OS):** O watcher do macOS/Linux (`inotify`) falha silenciosamente caso o file handle limit do sistema operacional seja extrapolado pelo número monstruoso de pastas. O Loam avisa quando a infraestrutura falha de subir todos os processos.
