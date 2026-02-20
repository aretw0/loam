# ADR 008: Strict Mode e Fidelidade (Consistent Typings)

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

A conversão entre Plain Text Formats varia conforme as implementações bibliotecárias na linguagem base (`encoding/json`, `yaml.v3`).
A representação de um ID gigantesco (`10382093820980193`) gerado como `int64` salva impecavelmente num ambiente, mas um leitor Go não-tipado padrão frequentemente decodifica tipos numéricos longos como inteiros truncados ou notações científicas destrutivas (`float64`) no unmarshaling genérico. Perder essa referência numérica quebra completamente o banco em aplicações escaláveis.

## Decisão

Para os programadores construindo sob o Loam utilizando mapas genéricos, criamos a opção arquitetural do **Strict Mode** injetável via configuração de Adapter.
Ativar o Strict Mode força o Loam aplicar uma rotina baseada em recursão (`recursiveNormalize` no pós-parsing) atenuando disparidades em parsers agnósticos e forçando que todos os números em transição ganhem fidelidade absoluta (`json.Number`), garantidos sobre Strings imutáveis sob o capô.

## Consequências

* **Pró (Consistência Poliglota):** O desenvolvedor salva um ID gigante em Markdown num dia e abre com um Client JSON noutro, validando os tipos numéricos sem perdas flutuantes (overflow).
* **Contra (Overhead e Idiomas):** Utilizar classes puras Go como `map[string]any` no desenvolvimento se afasta do ideal pois lidar com a formatação wrapper `json.Number` requer assertions mais burocráticos explicitamente. A performance recebe leve oneração (O(N) CPU overhead de normalização linear profunda nos mapas recebidos pela I/O de disco). Por conta deste contra, **é deixada inativa por default (Strict False) para uso ordinário de Go Idiomático.**
