# ADR 006: Smart CSV & Fidelidade de Dados Estruturados

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

Enquanto analisamos arquivos independentes e plain-text, o Loam foi desafiado a aceitar arquivos "Tabela-Cêntricos" (CSV) além de árvores "Documento-Cêntricas" (JSON/YAML/MD). Usuários pediam flexibilidade de analisar seus modelos no Excel/Pandas, mas os modelos do Loam permitem aninhamento de objetos genéricos complexos (Listas dentro de Mapas) (`Type Erasure`). Exportar esse dado para uma estrutura plana como um CSV causa obliteração estrutural, perdendo a "fidelidade do domínio": se carregarmos de volta, ele volta quebrado (tudo vira string sem profundidade).

## Decisão

Instituiu-se a convenção chamada de **Smart CSV Parsing**.
Ao interceptar chaves que denotem Slices ou Slices de Estruturas na serialização, o provedor CSV (`csv.Marshaler` proprietário do Loam) os encapsuliza como blocos validados (stringificados) de JSON dentro da própria célula CSV.

Na hora de leitura (Unmarshaling), o Loam inspeciona iterativamente a célula e executa a engenharia reversa para "desachatar" o documento à forma poliglota genérica de que a API Typada necessita para entregar ao `Service`.

## Consequências

* **Pró (Domínio Protegido):** Garantiu-se que migrar uma rota do Loam focada em JSON por uma focada em CSV fosse absolutamente livre de refatorações na struct das aplicações subjacentes. As abstrações sobrevivem intactas.
* **Pró (Consumo Misto):** Dados gerados podem ser analisados em rotinas Python (suportando decodificação de coluna) mesmo utilizando colunas JSON no meio e planilhas excel legíveis nos dados estáticos.
* **Contra (Limitação Heurística de Leitura):**  A inteligência reversa tenta adivinhar se `{...}` é json ou apenas texto que o usuário colocou chaves em volta por acaso. Sem `Strict Mode`, Strings ambíguas sofrem do problema nativo do unmarshal automático da Golang (onde a intenção "pura string" é perdida).
