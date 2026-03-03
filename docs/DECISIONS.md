# Strategic Decisions (Feb 2026)

> This document captures strategic decisions about Loam's positioning and future direction based on ecosystem analysis.

## Decision 1: Loam Maintains Autonomy with Emergent Synergies

**Date**: 2026-03-02  
**Decision**: Loam é e permanece uma engine genérica auto-suficiente. Sinergias com Trellis e outros projetos são emergentes, não forçadas.

**Rationale**:

- Loam foi projetado como engine embarcável para múltiplos use cases
- Trellis **escolheu** usar Loam porque features são genéricas e poderosas
- Forçar branding "Loam = parser do Trellis" limita percepção e adoção
- Outros projetos devem se sentir encorajados a usar Loam (PKM, config, ETL, etc.)

**Philosophical Principle**: "Autonomy with Emergent Synergies"

```
✅ Cada projeto (Lifecycle, Loam, Trellis) é auto-suficiente
✅ Integrações são escolhas naturais, não dependências estruturais
✅ Sinergia emerge do uso real, não é prescrita
❌ Não acoplar branding (Loam não é "o parser do Trellis")
❌ Não forçar roadmap baseado em um único stakeholder
```

**Impact**:

- README mantém posicionamento genérico
- ECOSYSTEM.md documenta Trellis como **um uso importante**, não o único
- Examples mostram standalone e integração com igual destaque
- Features avaliadas por utilidade genérica, não para um projeto específico

Veja [ADR 014: Understanding Trellis-Loam Symbiosis](architecture/014-understanding-trellis-loam-symbiosis.md) para detalhes e case study.

---

## Decision 2: Demand-Driven Adapter Strategy (No Speculative Adapters)

**Date**: 2026-02-28  
**Decision**: New adapters (HTTP, Redis, Database, S3, etc.) are implemented ONLY when there is real, persistent market demand. Not speculatively.

**Rationale**:

- FS + Git is perfectly optimized for Trellis workflows (80% of current use)
- Speculative adapters = maintenance burden with zero immediate value
- Architecture (Repository interface) is ready for new adapters when needed
- Can implement in <2 days when demand hits (proven with FS adapter)

**Implementation Framework**: 3-Test Decision Gate

1. **Is it Loam's responsibility?** (Or is it DALgo, client code, or framework concern?)
2. **Is there market pressure?** (Real users asking for it, not "would be cool")
3. **Is architecture ready?** (Can implement without major refactoring?)

**Adapter Roadmap**:

- **FS + Git** (v0.10): Critical, implemented ✓
- **HTTP** (v0.12+): Only if Arbour Phase 1 shows need
- **Redis** (v0.13+): Only if Life-DSL distributed workers demand it
- **Database** (Never): Use DALgo instead
- **S3** (Future): Only if cloud/SaaS variant of Trellis exists

**See**: [ADR 013: Demand-Driven Adapter Strategy](./architecture/013-demand-driven-adapter-strategy.md)

---

## Decision 3: FS + Git Specialization is Intentional

**Date**: 2026-03-02  
**Decision**: A especialização de Loam em FS + Git backend é uma escolha intencional alinhada com a filosofia "Everything as Code".

**Rationale**:

- Git audit trail é feature, não overhead
- Workflows, configs, documentos vivendo em version control é parte da filosofia
- Simplicidade > False Generality (não tentar ser tudo para todos)
- Loam **não é** (e não deve ser) um "database abstraction layer"

**What This Means**:

- ✅ FS+Git é a especialização correta de Loam
- ✅ Use cases que se beneficiam de Git versioning são ideais para Loam
- ✅ Para DB abstraction (SQL, NoSQL, Cloud), usar DALgo (camada diferente)
- ❌ Não adicionar backends especulativamente (HTTP, Redis, S3) sem demanda clara

**Trade-off Aceito**:

- Abrimos mão de "database abstraction" universal
- Ganhamos simplicidade, confiabilidade, e coerência filosófica

---

## Decision 4: Strict Mode + ReadOnly são Features Genéricas

**Date**: 2026-03-02  
**Decision**: Loam's Strict Mode (v0.10.4) e ReadOnly Mode (v0.10.6) são features genéricas úteis para múltiplos contextos, não exclusivas para Trellis.

**Why**:

**Strict Mode** (WithStrict(true)):

- Garante numeric types consistentes entre YAML, JSON, Markdown
- Previne "1000" ser float64 em YAML mas int em JSON
- Útil para **qualquer app** que precisa de type safety em comparações numéricas
- Trellis usa (workflow conditions), mas PKM apps, config parsers também se beneficiam

**ReadOnly Mode** (WithReadOnly(true)):

- Garante que a engine **nunca modifica** os arquivos source
- Modificações acontecem via Git commits explícitos
- Previne data loss ou corrupção acidental
- Útil para **qualquer app** que trata source files como read-only truth

**Implication**: Estas features são parte do core de Loam, não quirks para um projeto específico. Qualquer aplicação pode se beneficiar.

---

## Decision 5: Document Integrations Without Forcing Dependencies

**Date**: 2026-03-02  
**Decision**: Documentar casos de uso importantes (como Trellis) de forma clara, mas sem forçar acoplamento conceitual.

**Current State (Corrigido)**:

- README posiciona Loam como engine genérica
- ECOSYSTEM.md menciona Trellis como **um caso de uso importante entre vários**
- Examples incluem standalone use e integrations com igual destaque
- Trellis integration não é "escondida", mas também não domina a narrativa

**Required Documentation**:

- [x] README: Manter posicionamento genérico
- [x] ECOSYSTEM.md: Clarificar que sinergia é emergente
- [x] RECIPES.md: Documentar integration patterns (5 recipes: Workflow, PKM, Config, ETL, CLI)
- [x] Code comments: Explicar features sem mencionar projetos específicos

---

## Decision 6: DALgo is Complementary, Not Competitive

**Date**: 2026-02-28  
**Decision**: Loam and DALgo solve different problems at different layers. They are complementary, not competing.

**Positioning**:

- **Loam**: "How do we load workflow definitions?" (Parser layer)
- **DALgo**: "How do we abstract data access?" (Persistence layer)

**Use Case Separation**:

```
Scenario: Trellis task needs to read data from SQL + Firebase + Local files

Solution:
- Trellis loads workflow definition via Loam (FS+Git)
- Inside "read_customer_data" task:
  - DALgo used internally to abstract DB access
  - Loam not involved
```

**Why Not Combine?**:

- Loam is optimized for Git-backed documents (trees, workflows)
- DALgo is optimized for relational/document DB abstraction
- Combining would muddy both (neither is ideal for the other's use case)

---

## Decision 7: Version Independence with Communication

**Date**: 2026-03-02  
**Decision**: Loam mantém versionamento independente. Comunicação com projetos dependentes (Trellis, etc.) acontece para breaking changes, mas não há "stability pact" forçado.

**Philosophy**:

- Loam evolui baseado em suas próprias necessidades e feedback do ecossistema
- Breaking changes são comunicados com antecedência para projetos conhecidos
- Mas nenhum projeto "trava" o roadmap de Loam
- Migration guides fornecidos quando necessário

**How It Works**:

```
Loam v0.10.x (Stable Document Engine)
  ↓ (usado por)
Trellis v0.7.x (integra via interface pública de Loam)
  ↓
Se Loam v0.11 introduz breaking changes:
  → Aviso prévio para Trellis maintainers
  → Migration guide publicado
  → Trellis decide quando atualizar (não é bloqueante)
```

**Key Principle**: Autonomia com comunicação, não dependência estrutural.

---

## Review Schedule

These decisions should be reviewed:

- **Quarterly**: Check if adapter demand has changed
- **Major release**: Reconsider positioning if ecosystem shifts significantly
- **When Trellis phases change**: Align adapter/feature priorities

Next review: **May 2026** (after Arbour Phase 1 MVP + Life-DSL Phase 2b)
