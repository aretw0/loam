# ADR 004: Modelo de Operações Transacionais

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

Salvar dez arquivos Markdown independentes diretamente no disco abre uma janela perigosa: O processo falha no meio e o estado fica instável ou quebrado – com metade dos arquivos atualizados de uma forma e os outros desatualizados (corrupção parcial).
Para simular um verdadeiro "Database" no sistema de arquivos, o Loam exige garantias de integridade análogas aos modelos ACID base.

## Decisão

Adotamos a premissa de que a escrita será isolada utilizando fluxos de **Batch Transactions (Modelos Atômicos)**. Toda alteração no estado deve ocorrer dentro de um contexto ou escopo gerido e aprovado pelo Adaptador de File System antes que qualquer persistência final seja firmada.

### Mecânica Central

1. O cliente chama o pipeline instanciado em `WithTransaction(fn)`.
2. O Adapter abre a transação, rastreando virtualmente as alterações exigidas (`Staged`). As operações executam *in-memory* ou em *scratch spaces* antes de colidirem.
3. Se todo o lote obtiver sucesso (retorno `nil`), executa-se a Escrita em Lote.
4. Finaliza-se encadeando um commit semântico do Git agrupando todas as escritas do lote em uma única hash SHA256 no histórico de versões para garantir não apenas a atomização em disco, como na auditoria.

## Consequências

* **Pró (Garantia de Estado):** Fim das quebras silênciosas das informações. Menos instabilidades geradas por processos abruptamente cancelados (ex: falhas de energia/SIGKILLs).
* **Contra (Não é Thread-safe isoladamente):** Estruturas sequenciais na transação não são protegidas inerentemente caso paralelizadas dentro de uma interface genérica de acesso, requerendo abordagens controladas na manipulação dos pipelines.
* **Contra (Memória):** Para lotes gigantescos (`Batch Size` > dez milhares), o consumo de RAM poderá elevar devido ao stage-in-memory.
