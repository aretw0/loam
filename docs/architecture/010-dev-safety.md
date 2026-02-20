# ADR 010: Dev Safety (Safety by Default) e Read-Only Sandboxing

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

Como atuamos diretamente no *File System* do usuário, diferentemente de um banco SQL que roda num container alheio `localhost:5432`, o risco do desenvolvedor escrever um `.yaml` mal-formado durante o Loop de `run/test` e poluir/vazar dados da raiz de produção (ex: os próprios arquivos do desenvolvedor gerindo o repositório em desenvolvimento) é drástico. Testes unitários rodando em scripts e esvaziando a pasta de projetos seria letal para a confiança na biblioteca.

## Decisão

A adoção de uma camada severa de interceptação de Paths na orquestração:

1. **Sandbox Default (`DevSafety` ON by Default)**: Se o ambiente for detectado como desenvolvimento (`go run`, `go test`), o `ResolveVaultPath` jamais gravará no *Vault* local (passado via config). Ele construirá ativamente um cache descartável (`%TEMP%/loam-dev/`) e usará-o silenciosamente.
2. **Explicitamente Inseguro**: O desenvolvedor deve assumir o risco explícito colocando `WithDevSafety(false)` para escrever em disco fora de binários compilados em release.
3. **Strict Read-Only (`WithReadOnly(true)`)**: Além da Sandbox, foi instituído no núcleo que motores puramente visuais (como as extensões e portais do Trellis) possam operar nas pastas originais de desenvolvimento desde que atestem ser *Read-Only*, onde o Adaptador bloqueia *hardcoded* qualquer `.Save()`, `.Delete()` e `.Sync()`, retornando o erro nativo fixado `ErrReadOnly`.

## Consequências

* **Pró (Trust):** Bibliotecas perigosas (Write-to-folder) nunca destroem projetos em execuções de primeira viagem de contribuintes novos.
* **Pró (Tooling Acoplado):** Extensões como editores visuais do Trellis podem explorar e parsear Vaults em produção no terminal sem receio de sujar a Stage tree do Git do usuário.
* **Contra (Ergonomia Inicial/Frustração):** Atores de primeira viagem frequentemente criam o script e não entendem porquê nenhum arquivo está sendo gerido na sua pasta de execução (por não lerem sobre Sandbox), pensando ser um Bug do Loam até descobrir a flag do Override.
