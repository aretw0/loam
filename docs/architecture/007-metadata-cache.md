# ADR 007: Indexação Persistente (Metadata Cache)

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

Trabalhar exclusivamente nos domínios do Sistema Operacional rodando leituras (`WalkDir`) iterativas para consultas como listar "10.000 cartões do kanban para renderizar a página inicial" é custoso.
Com metadados profundos, ler e parsear 10.000 JSONs a cada consulta causaria penalidade de tempo excessivamente oneroso por ser bloqueante em cenários sincrônicos no cliente final.

## Decisão

Adotamos o conceito de uma estante central otimizada com a construção do **Metadata Cache (Index Persistente)**.
O *FS Adapter* foi estendido para manter silenciosamente um arquivo consolidado em `.loam/index.json`. Toda modificação/escrita/deleção (`Save()`, `Delete()`) que ocorre através do Loam, e todos os retornos assíncronos do Watcher informam o FS Adapter que a indexação tem atualizações pendentes. Esse index cache central armazena chaves rápidas sobre os Caminhos de Disco para atalho computacional durante comandos exploratórios listados.

## Consequências

* **Pró (Eficiência em Alta Escala):** Buscas que escaneariam Múltiplos Megabytes demorando milhares de milissegundos passam a custar um parse direto em memórias baixando as queries a tempos na casa da centena ou dezena (`±13ms`).
* **Contra (Dissonância em Coleções/Arrays Arquivais):** O indexador trabalha mapeando metadados atrelados à **rota de arquivo em disco**. Se um documento .json abriga múltiplos perfis (ex: um Root Array `[{a}, {b}]`), o indexador atual não abre esse sub-nível e o `Metadata Cache` não trará benefícios drásticos sob essa topologia de estruturação de dados pelo usuário.
* **Risco de Sync Stale Cache:** Existe o risco pontual de colisão nos registros persistidos, necessitando de Reconciliações severas ao reiniciar o programa para evitar que a tabela de indexação mostre objetos no disco que não estão lá.
