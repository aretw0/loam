# ADR 011: Interfaces de Capacidade Padrão (Capability Interfaces)

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

Conforme o Loam cresceu, funções como `Watch()` (Reatividade de FS), `Sync()` (Push/Pull do Git), e `Reconcile()` (Cold Start Repair) não faziam sentido e quebravam contratos rígidos para adaptadores conceituais (ex: um Adapter SQL não possui conceito de Sync do Git; um adaptador estático Mockado de testes não monitora OS events). Forçar essas funções no contrato Core de um `Repository` gerava inchaço da arquitetura (*Fat Interface*).

## Decisão

Seguindo os preceitos clássicos da interface Go (Pequenas Peças e Segregação de Interfaces - ISP), as capacidades "Avançadas" devem ser fornecidas via sub-interfaces:

- `Watchable` (Exige método Watch)
- `Syncable` (Exige método Sync ao Git)
- `Reconcilable` (Existem método de Reconciliar)

Dessa forma, a implementação do motor lógico de alto nível (`core.Service`) utilizará um **Runtime Check (*Type Assertion*)** para descobrir quais poderes o adaptador acoplado pelo usuário suporta.

## Consequências

- **Pró (Minimalismo e Breaking Changes):** Novos Adapters podem ser escritos pela comunidade para plugar no Loam sem precisar prover funções complexas e indesejadas que não condigam com o cenário adotado.
- **Pró (Degradação Graciosa):** O aplicativo continua rodando com falhas limpas operacionais. (ex: chamar `Service.Sync()` com adaptador `In-Memory` retorna imediatamente um `CapabilityNotSupported` explícito que o client-code pode ignorar).
- **Contra (Não Descoberto em Tempo de Compilação):** Erros de arquitetura só ocorrem em tempo de execução (*Runtime*), pois o fato de um Adapter falhar na checagem Type Assertion nunca resultará no travamento (*panic*) da fase de compilação da pipeline Go.
