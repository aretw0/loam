# Arquitetura Técnica

O **Loam** adota uma **Arquitetura Hexagonal (Ports & Adapters)** para garantir desacoplamento, testabilidade e extensibilidade. O núcleo do sistema opera independentemente do mecanismo de persistência, permitindo trocar o sistema de arquivos (FS + Git) por outras implementações (SQL, API, etc.) sem alterar as regras de negócio.

## Tabela de Conteúdo

- [Visão Geral](#visão-geral)
- [Fluxos de Execução](#fluxos-de-execução)
- [Componentes](#componentes)
- [Decisões Arquiteturais Chave](#decisões-arquiteturais-chave)
- [Estratégia de Testes](#estratégia-de-testes)

## Visão Geral

```mermaid
graph TD
    subgraph "External World (Primary Adapters)"
        CLI[CLI / Command Line]
        App[Your Go App]
    end

    subgraph "Parsing Logic (Agnostic)"
        PD[ParseDocument]
        SD[SerializeDocument]
    end

    subgraph "Hexagon (Core Domain)"
        direction TB
        Service[Core Service - Business Rules]
        Model[Domain Model - Document / Metadata]
    end

    subgraph "Infrastructure (Secondary Adapters)"
        FSAdapter[FS Adapter - File persist]
        Git[Git Client]
        Cache[Index Cache]
    end

    CLI -->|Calls| Service
    App -->|Calls| Service
    
    Service -->|Uses| Model
    Service -->|Port Interface| FSAdapter
    
    FSAdapter -->|Impl| Git
    FSAdapter -->|Impl| Cache
```

## Fluxos de Execução

### Batch Transaction (Atomicidade)

O fluxo de uma transação em lote garante que arquivos só sejam persistidos e commitados se todas as operações tiverem sucesso.

```mermaid
sequenceDiagram
    participant Client
    participant Service
    participant Adapter
    participant Git

    Client->>Service: WithTransaction(fn)
    Service->>Adapter: Begin()
    Adapter-->>Service: Transaction (tx)
    
    Service->>Client: fn(tx)
    
    Note right of Client: Operações em Memória
    Client->>+tx: Save(doc1)
    tx-->>-Client: ok (staged)
    Client->>+tx: Save(doc2)
    tx-->>-Client: ok (staged)
    
    Client-->>Service: return nil (Success)
    
    Note right of Service: Persistência Atômica
    Service->>+tx: Commit("feat: update docs")
    tx->>Adapter: Write Files (Disk)
    tx->>Git: git add .
    tx->>Git: git commit -m "feat:..."
    Git-->>tx: ok
    tx-->>-Service: ok
    
    Service-->>Client: ok
```

### Ciclo de Vida do Documento

```mermaid
stateDiagram-v2
    [*] --> Memory : New()
    Memory --> Staged : Save()
    Staged --> Persisted : Commit()
    Persisted --> [*]
    
    state "In Context (Dirty)" as Memory
    state "Filesystem (Pending)" as Staged
    state "Git History (Safe)" as Persisted
```

## Componentes

A estrutura de diretórios do projeto reflete diretamente a arquitetura hexagonal adotada:

```text
.
├── cmd/                # Pontos de entrada (CLI)
├── internal/
│   └── platform/       # Implementações de Infraestrutura (Hidden)
├── pkg/
│   ├── adapters/       # Adaptadores (FS, Git) - "Fora do Hexágono"
│   ├── core/           # Domínio e Portas - "Dentro do Hexágono"
│   └── typed/          # Camada de Tipagem (Generic Wrapper)
├── loam.go             # Facade pública
└── examples/           # Exemplos de uso
```

### 1. Core Domain (`pkg/core`)

- **Entidades**: `Document` (ID, Content, Metadata).
- **Ports (Interfaces)**: `Repository` (Save, Get, List, Delete) — O "Abstract Librarian".
- **Services**: `Service` (Orquestra validações e chama o Repository).
- **Dependências**: Zero dependências de infraestrutura.

### 2. Adapters (`pkg/adapters`)

Implementações concretas dos Ports definidos no Core.

- **FS Adapter (`pkg/adapters/fs`)**: Implementa `core.Repository`. A "Estante Física".
  - Gerencia a **persistência física** de Documentos (serializando como Markdown/YAML/JSON).
  - Utiliza `pkg/git` para controle de versão.
  - Mantém um **Cache (.loam/index.json)** para listagens rápidas (Otimização).

### 3. Typed Layer (`pkg/typed`)

Camada de conveniência que envolve o `core.Service` para fornecer APIs genéricas (`TypedRepository[T]`).

- **Responsabilidade**: Marshaling/Unmarshaling de structs Go para `core.Metadata` (map[string]any).
- **Benefício**: Garante que o consumidor trabalhe com tipos fortes, enquanto o core permanece dinâmico.

### 4. Internal Platform (`internal/platform`)

O "chão de fábrica" do sistema. Contém a implementação concreta do bootstrap e segurança.

- **Responsabilidade**: Factory method, Configuração de Opções, Sanitização de Paths (`ResolveVaultPath`) e utilitários de Dev (`IsDevRun`).
- **Visibilidade**: Privado para a biblioteca, não importável externamente.

### 5. Public Facade (`github.com/aretw0/loam`)

A fachada pública que simplifica o uso da biblioteca.

- **`loam.go`**: Expõe aliases para as funções do `internal/platform` (`New`, `Init`, `Sync`, `Option`) e `pkg/typed` (`OpenTyped`).
- **Objetivo**: Manter a raiz do projeto limpa e fornecer uma API estável enquanto a implementação evolui internamente.

## Decisões Arquiteturais Chave

### 1. Storage Engine: Filesystem + Git (Driver Padrão)

- **Formato:** Arquivos de texto (`.md`, `.json`, `.yaml`, `.csv`) gerenciados pelo FS Adapter.
- **Smart Retrieval (Fuzzy Lookup):** Ao buscar um documento sem extensão, o adapter escaneia o diretório por extensões suportadas.
- **Transações:** O Git atua como *Write-Ahead Log*.
- **Smart Gitless:** O sistema detecta automaticamente se deve usar Git. Se `.git` não existir, mas `.loam` (system dir) existir, ele opera em modo "Gitless" (apenas FS), permitindo uso flexível em ambientes contêinerizados ou efêmeros.
- **Semântica de Commit:** O adapter `fs` lê `commit_message` do `context.Context` (se Git estiver ativo).

### 2. Cache de Metadados

- **Problema:** Listar milhares de arquivos lendo do disco é lento.
- **Solução:** Index Persistente (`.loam/index.json`) mantido pelo Adapter FS. O diretório (`.loam`) é configurável via `WithSystemDir`.
- **Invalidação:** Baseada em timestamp (`mtime`).
- **Performance:** Reduz tempo de listagem de segundos para milissegundos (ex: 6s -> 13ms para 1k documentos).

### 3. Segurança (Dev Safety)

- **Isolamento**: Em modo de desenvolvimento (`go run`, `go test`), o Loam redireciona automaticamente operações para um diretório temporário (`%TEMP%/loam-dev/`) para evitar sujar o repositório do usuário.
- **ForceTemp**: Configurável via `loam.Config`.

## Estratégia de Testes

### 1. Unitários (`pkg/core`)

Testam regras de negócio usando Mocks em memória. Execução instantânea.

### 2. Integração (Integration Tests)

Testam o fluxo completo (Service + FS Adapter + Git) em diretórios temporários reais. Garantem que o contrato de persistência é cumprido.

### 3. Benchmarks (`examples/benchmarks`)

Medem a performance do Adapter e eficácia do Cache.
