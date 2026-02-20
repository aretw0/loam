# ADR 003: Armazenamento em Plain Text First

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

As aplicações hoje possuem grandes necessidades de escalonamento que, em bancos NoSQL ou de Grafos, costumam ser supridas utilizando formatos binários condensados.
Adotar um banco de dados tradicional agilizaria muito as buscas e junções (join) no Loam, mas a opacidade desses binários iria ferir diretamente a intenção inicial de ter informações perfeitamente legíveis para o ser humano ("Human Readable").

## Decisão

Foi adotado que o Loam operará focado no paradigma de **Plain Text First (Texto Puro)**. Como premissa, o Loam só manipulará tipos abertos textuais (como o **Markdown**, e para contextos padronizados **JSON**, **YAML** ou arquivos de configuração **CSV**).

Toda manipulação de sistema passa por parsers configuráveis (Interfaces `Serializer`). Ao embutir a estrutura na máquina host, se ela desinstalar o sistema hospedeiro o cliente final abrirá os dados no Bloco de Notas sem ter reféns.

## Consequências

* **Pró (Interoperabilidade Absoluta):** Funciona out-of-the-box com o ecossistema atual de editores (VSCode, Notepad++, Obsidian, Neovim).
* **Pró (Portabilidade):** Migração para qualquer infraestrutura é "copy and paste" direto. Extensibilidade alta de análise de texto padrão usando GNU/Linux utils (`grep`, `awk`).
* **Contra (Custo de Parsing Repetido):** Em cada ciclo de vida do FS Adapter, arquivos de texto necessitam de alocações na memória durante operações de Marshaling e Unmarshaling (I/O).
* **Mitigação (Indexação/Cache):** Como contrapeso fundamental à escolha do texto puro, desenvolvemos a indexação híbrida de "Metadata Cache", descrita posteriormente em uma ADR sobre a estratégia do `.loam/index.json`.
