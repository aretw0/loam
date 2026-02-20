# ADR 001: Arquitetura Hexagonal & Event-Driven

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

Um motor de documentos destinado a servir como biblioteca para diversas categorias de aplicações (gestão de configuração, ferramentas PKM, pipelines de dados) exige alta coesão e baixo acoplamento. Precisávamos que o núcleo do sistema (`Core`) operasse independentemente da forma como os arquivos são lidos, salvos ou rastreados no sistema operacional.
Ao mesmo tempo, as aplicações clientes do Loam precisam reagir a mudanças no conteúdo (como uma edição feita pelo usuário diretamente via VSCode) de forma síncrona.

## Decisão

O Loam adotou um modelo arquitetural híbrido:

1. **Arquitetura Hexagonal (Ports & Adapters)** para o **Write Path**: Garantindo que a lógica de aplicação e o modelo de domínio (o Hexágono) comuniquem-se com o mundo externo apenas por interfaces (`Ports`). A persitência de Arquivos (`FS`) e o `Git` são meros *Adapters* (Secondary Adapters), enquanto a `CLI` e as `Aplicações Go` que o instanciam são *Primary Adapters*.
2. **Event-Driven Architecture** para o **Read/Reactive Path**: Garantindo reatividade assíncrona. O *Adapter FS* monitora o sistema operacional, mapeia os eventos de arquivo (raw events) em eventos de domínio (`Domain Events`), e os transmite ao *Core Service* através de um `Event Broker` com *buffers* nativos (Go channels).

## Consequências

* **Pró (Testabilidade):** O core e o modelo transacional podem ser testados com mocks (estratégia `In-Memory`) instantaneamente, sem encostar em I/O.
* **Pró (Plataforma Aberta):** A arquitetura base suportaria novos *Adapters* (como SQL, S3, ou redes P2P nativas via CRDTs) no futuro sem refatorar o serviço principal.
* **Contra (Overhead Teórico):**  Maior complexidade gerencial nas goroutines e canais de comunicação para garantir *Graceful Shutdown* e gerenciar estados de corrida (Race Conditions) ao repassar os eventos para os clientes.
