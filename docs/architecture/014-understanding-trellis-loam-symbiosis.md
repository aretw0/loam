# ADR 014: Understanding Trellis-Loam Symbiosis (Autonomy with Emergent Synergy)

**Data:** 2026-03-02  
**Status:** Aceita

## Contexto

Loam foi desenvolvido como engine genérica e auto-suficiente para persistência de conteúdo e metadados com Git audit trail. **Este design permanece correto.**

Em fevereiro/2026, análise do ecossistema revelou **sinergia emergente profunda com Trellis**:

1. **Trellis escolheu Loam como built-in parser centralizado**
   - `trellis.New(repoPath)` inicializa `loam.Init()` por padrão
   - `pkg/adapters/loam/loader.go` converte documentos Loam em nós Trellis (~500 linhas)
   - Workflows em YAML/JSON/Markdown carregados via Loam

2. **Features de Loam são valiosas para Trellis (mas não exclusivas)**
   - Strict Mode (v0.10.4) → Type consistency (útil para qualquer app que precisa de types rigorosos)
   - ReadOnly Mode (v0.10.6) → Engine safety (útil para qualquer app que não deve modificar source)
   - Watch support → Hot reload (útil para qualquer app que reage a mudanças)

3. **Trellis demonstra poder de Loam sem defini-lo**
   - Loam é genérico: útil para PKM, config management, ETL, workflows
   - Qualquer projeto pode integrar sem acoplamento
   - Integração com Trellis valida design, não o limita a esse caso de uso

## Decisão

### Posição de Loam: "Ferramenta Autônoma com Casos de Uso Emergentes"

```
Loam Identity:
  "Embedded reactive & transactional engine for content & metadata"
  • Multi-format document persistence (YAML/JSON/Markdown/CSV)
  • Git audit trail built-in
  • Type-safe metadata extraction
  • Reactive watchers
  • Transactional batches

Casos de Uso (todos válidos):
  ✓ Workflows (Trellis — integração emergente, production-proven)
  ✓ PKM assistants (storage layer genérico)
  ✓ Configuration management (GitOps patterns)
  ✓ Local data pipelines (ETL with Git versioning)
  
Princípio: Autonomia com sinergia emergente. Cada projeto é auto-suficiente.
```

### Implicações Arquiteturais

1. **Loam mantém autonomia e não prioriza um único caso de uso**

   ```
   Do:
     ✅ Multi-format support (valioso para workflows, configs, PKM)
     ✅ Metadata extraction (útil para qualquer estrutura de dados)
     ✅ Transactional writes (integrity em qualquer contexto)
     ✅ Watch support (reatividade para qualquer app)
     ✅ Git audit trail (versioning universal)
   
   Don't:
     ❌ Acoplar features específicas de Trellis no core
     ❌ Forçar branding atrelado a um projeto
     ❌ Priorizar roadmap baseado em um único stakeholder
   ```

2. **FS + Git como especialização intencional, não limitação**

   ```
   Why:
     • "Everything as Code" philosophy
     • Audit trail é feature, não overhead
     • Simplicidade > False Generality
   
   Adapters:
     • Implementar novos quando houver demanda de mercado (não especulativo)
     • Decisão via 3-Test Framework (ADR 013)
     • Arquitetura pronta (Repository interface permite extensão)
   ```

3. **Documentar integrações sem acoplamento conceitual**
   - ECOSYSTEM.md reflete sinergia emergente
   - Examples incluem Trellis integration como **um entre vários casos de uso**
   - README mantém positioning genérico

## How This Affects Product & Roadmap

### Documentation Philosophy

**Princípio**: Manter autonomia de cada projeto enquanto documenta sinergias emergentes.

```markdown
README Positioning:
  "Embedded reactive & transactional engine for content & metadata"
  
  Ideal para toolmakers que constroem:
  • PKM assistants (storage layer)
  • Configuration management (GitOps patterns)
  • Local data pipelines (ETL with versioning)
  • Workflow engines (e.g., Trellis uses Loam for multi-format parsing)

ECOSYSTEM.md:
  "Loam é usado por Trellis como parser centralizado de workflows"
  "Esta integração demonstra a flexibilidade do design, não o limita"
  
Examples:
  • examples/basics/ → Standalone usage
  • Trellis integration detailed in [trellis repo](https://github.com/aretw0/trellis)
```

### Adapter Strategy

Ver [ADR 013: Demand-Driven Adapter Strategy](./013-demand-driven-adapter-strategy.md)

Adapters são implementados apenas quando **há demanda real de mercado**, seguindo 3-Test Framework:

1. É responsabilidade de Loam? (ou DALgo, client code, framework?)
2. Há pressão de mercado? (usuários pedindo, não "seria legal")
3. Arquitetura está pronta? (implementa sem refactoring major?)

**Não**: "Implementar HTTP porque Trellis pode precisar"  
**Sim**: "Implementar HTTP quando Arbour Phase 1 mostrar necessidade clara"

### Version Independence

Loam e Trellis mantêm versionamento independente com comunicação sobre breaking changes:

```
Loam v0.10.x (Stable Document Engine)
  ↓ (usado por)
Trellis v0.7.x (Stable Workflow Engine)
  ↓ (usado por)
Arbour (Production App)

Breaking changes:
  • Loam → Avisa Trellis maintainers antes
  • Trellis → Não depende de internals de Loam (usa interface pública)
  • Migration guides quando necessário
```

## Trellis Integration Case Study

### Como Trellis Integra Loam

Ver [trellis.go](https://github.com/aretw0/trellis/blob/main/trellis.go) na repo de Trellis para implementação completa.

```go
// trellis.go excerpt
func New(repoPath string, opts ...Option) (*Engine, error) {
    // Initialize Loam with strict + readonly
    repo, err := loam.Init(absPath,
        loam.WithStrict(true),        // Numeric type consistency
        loam.WithReadOnly(true),      // Engine never modifies source
    )
    
    // Create typed repository for workflow metadata
    typedRepo := loam.NewTypedRepository[loamAdapter.NodeMetadata](repo)
    
    // Use Loam adapter to convert Doc → Node
    eng.loader = loamAdapter.New(typedRepo)
    
    return eng, nil
}
```

### Features de Loam Úteis para Trellis (e Outros Projetos)

| Feature | Valor para Trellis | Valor Genérico |
|---------|-------------------|----------------|
| **Strict Mode** | Type safety em conditions | Qualquer app que precisa numeric consistency |
| **ReadOnly** | Previne modificações acidentais | Qualquer app que trata source como imutável |
| **Watch** | Hot reload de workflows | Qualquer app que reage a mudanças de arquivos |
| **Markdown parsing** | Workflows em .md | PKM, docs, blogs, wikis |
| **Metadata extraction** | Node configs via frontmatter | Qualquer estrutura de dados com metadata |
| **Transactional writes** | Safety durante batch updates | Qualquer app que precisa de atomicidade |

**Takeaway**: Trellis usa Loam porque features são genéricas e poderosas, não porque foram feitas exclusivamente para ele.

## Risks & Mitigations

### Risk 1: Acoplamento conceitual forçado

**Sintoma**: Branding de Loam se torna "o parser do Trellis", limitando percepção de uso genérico  
**Mitigação**:

- Manter README focado em features genéricas
- Documentar Trellis como integração importante, não exclusiva
- Exemplos standalone tão visíveis quanto integração
- Aplicar "autonomia com sinergia emergente" consistentemente

### Risk 2: Feature creep para agradar um único stakeholder

**Sintoma**: Features úteis apenas para Trellis entram no core de Loam  
**Mitigação**:

- Pergunta-chave: "Esta feature é útil fora do contexto de Trellis?"
- Se sim → Core (ex: `WithStrict` é genérico)
- Se não → Adapter layer (`trellis/pkg/adapters/loam/`)
- Code review com perspectiva de outros use cases

### Risk 3: Projetos se sentem "second class citizens"

**Sintoma**: PKM assistants, ETL tools sentem que Loam não é para eles  
**Mitigação**:

- Documentação clara: "Loam é genérico, Trellis é um usuário importante entre vários"
- Examples de standalone uso (PKM, config, ETL) mantidos atualizados
- Não usar terminologia "primary/secondary use cases"

## Relationship to Other Components (Autonomy with Integration)

```
┌───────────────────────────────────────────────────────┐
│ Lifecycle (Foundation Layer)                          │
│ • Signal handling                                     │
│ • Process management                                  │
│ • Event router                                        │
│ • Auto-suficiente: usado por Loam, Trellis, Arbour   │
└───────────────────────────────────────────────────────┘
                           │
        ┌──────────────────┴──────────────────┐
        ▼                                      ▼
┌──────────────────────┐          ┌──────────────────────┐
│ Loam                 │          │ Outros Projetos      │
│ (Document Engine)    │          │ (DB access, etc)     │
│                      │          │                      │
│ • Multi-format       │          │ • DALgo              │
│ • Git audit trail    │          │ • Outros adapters    │
│ • Metadata           │          │                      │
│ • Auto-suficiente    │          │                      │
└──────────┬───────────┘          └──────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────┐
│ Trellis (Workflow Engine)                            │
│ • State machine execution                            │
│ • Workflow routing                                   │
│ • Tool invocation                                    │
│ • **Escolhe** usar Loam (não depende estruturalmente)│
└──────────────────────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────┐
│ Arbour (Production App)                              │
│ • Business logic                                     │
│ • Usa Trellis (que usa Loam)                         │
└──────────────────────────────────────────────────────┘
```

**Filosofia**: Cada layer é auto-suficiente. Integrações são escolhas, não dependências estruturais forçadas.

- **Lifecycle**: Fundação para tudo (signal handling, workers)
- **Loam**: Engine genérica que **pode** ser usada por workflow engines
- **Trellis**: Workflow engine que **escolheu** Loam pela flexibilidade
- **Arbour**: App que usa Trellis (e transitivamente Loam)

Outros projetos podem usar:

- Lifecycle sem Loam
- Loam sem Trellis
- Trellis com outro parser (se implementarem adapter)

## Conclusão

**Loam é uma engine genérica com sinergia emergente demonstrada, mas não definida por nenhum único projeto.**

✅ **Autonomia**: Uso standalone é válido e encorajado  
✅ **Sinergia**: Trellis escolheu Loam (escolha natural, não forçada)  
✅ **Features genéricas**: Strict, ReadOnly, Watch úteis em múltiplos contextos  
✅ **Simplicidade**: FS+Git é escolha intencional alinhada com filosofia  
✅ **Roadmap**: Orientado por demanda real, não especulação  

**Princípio**: Autonomia com sinergia emergente. Cada projeto é auto-suficiente.

## See Also

- [ADR 013: Demand-Driven Adapter Strategy](./013-demand-driven-adapter-strategy.md)
- [PLANNING.md Phase 0.11: Ecosystem Positioning](../PLANNING.md#fase-0110-ecosystem-positioning)
- [ECOSYSTEM.md: Loam Role](../ECOSYSTEM.md)
- [trellis/trellis.go](https://github.com/aretw0/trellis/blob/main/trellis.go) — Integration entry point (Trellis repo)
- [trellis/pkg/adapters/loam/](https://github.com/aretw0/trellis/tree/main/pkg/adapters/loam) — Conversion logic (Trellis repo)
