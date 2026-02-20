# ADR 009: Content Extraction e Tratamento de Generic Data

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

A estrutura de dados das aplicações de gerenciamento de conteúdo (`Headless CMS` ou editores PKM) diverge historicamente de sistemas de banco de dados puramente seriais.
Em um CMS que gera sites, o "Corpo Secundário" do texto (como a área após o `---` Frontmatter yaml) é semanticamente diferente dos *Metadados* puros. Todavia, como construiríamos um wrapper forte em Go (`TypedLayer`) que pudesse lidar tanto com o caso de "Sou um gerador de blogs com conteúdo explícito" quanto "Sou um analisador de logs genérico focado estritamente em propriedades chave-valor JSON"?

## Decisão

Instituiu-se configurações bifurcadas nativas na estrutura do documento e parse:

1. **Formato Nativo Base**: As aplicações interagem através de uma abstração em que todo documento possui: `ID (string)`, `Content (string)` e `Metadata (map[string]any)`.
2. **Extraction ON (Default CMS-like)**: Ao ler um Markdown ou arquivo com chave "content" no JSON, o Parser isola isso na propriedade base `Document.Content`. Deixa os Metadados limpos apenas com tags auxiliares.
3. **Extraction OFF (Generic Data)**: Para preservação 1:1 rigorosa, o *payload inteiro* de um arquivo flui forçadamente para dentro de `Document.Metadata`. Assim, nenhum parser destrutivo roubará os campos (como roubar a key genérica `content` de um JSON complexo que não seja um blogpost). O desenvolvedor configura com `WithContentExtraction(false)`.

## Consequências

* **Pró (Agnosticismo de Domínio):** O `TypedRepository` consegue converter fielmente uma struct que possui `Fields` que conflitam com as definições internas do parser.
* **Pró (Zero Perda de Payload):** Documentos recebidos via APIs podem transitar e ser salvos no OS transparentemente com Extração Desligada, mantendo o envelope limpo.
* **Contra (Confusão Paramétrica):** Desenvolvedores implementando rotas genéricas podem ser pegos de surpresa por default e terem campos nomeados de `content` "sumindo" para dentro da raiz da struct `Document(ID, Content)`. Exige leitura minuciosa da documentação da Extraction.
