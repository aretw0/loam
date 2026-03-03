# ADR 013: Demand-Driven Adapter Strategy

**Data:** 2026-03-02  
**Status:** Aceita

## Contexto

Loam foi arquitetado com Hexagonal Pattern (ADR 001) para suportar múltiplos adapters (FS, HTTP, Database, Redis, etc). Historicamente, apenas o adapter **FS + Git** foi implementado, porque:

1. **Trellis** e **Life-DSL** demonstram caso de uso crítico para FS + Git (carregar workflows, configs)
2. **Arbour** e outros projetos no ecossistema também preferem FS + Git
3. Ninguém pediu por HTTP, Redis, S3, ou adapters de BD com pressão real de mercado

Essa seção documenta **quando e por quê** implementar novos adapters de forma responsável.

## Decisão

### Princípio: "Do One Thing Well" (Unix Way)

Adapters adicionais **NÃO** devem ser implementados especulativamente. Novos adapters devem ser implementados apenas quando:

1. **Há demanda real e persistente** (não uma "seria legal se...")
2. **O problema é apropriado para Loam** (não para camadas superiores ou DALgo)
3. **Há clareza sobre o caso de uso** (não especulação teórica)

### Critério de Decisão: 3 Testes

Antes de implementar um novo adapter, responda:

#### **Teste 1: É responsabilidade de Loam?**

```
Pergunta: "O cliente deveria carregar dados e passar para Loam, 
ou Loam deveria carregar diretamente?"

Sim para Loam adapter:
  ✅ "Frameworks precisam carregar workflows de múltiplas fontes"
  ✅ "Sistema requer hot reload de múltiplos backends"
  ✅ "Abstração é consistente com Repository interface"

Não para Loam adapter (responsabilidade de cliente/DALgo):
  ❌ "Preciso abstrair múltiplos bancos de dados"
     → Use DALgo em vez disso
  ❌ "Minha aplicação busca dados de HTTP"
     → Seu código faz HTTP call, passa dados para Loam
  ❌ "Quero carregar workflows, mas também dados de BD"
     → Split: Loam para workflows, DALgo para dados internos
```

#### **Teste 2: Há pressão do mercado?**

```
Seu critério: "Alguém realmente está usando Loam 
e pedindo por esse adapter?"

Forte indicador (implementar):
  ✅ Arbour Phase 1 MVP mostra que HTTP load é gargalo
  ✅ 3+ usuários independentes pedindo feature
  ✅ Projeto no ecossistema bloqueado sem isso

Fraco indicador (não implementar):
  ❌ "Seria legal ter HTTP support"
  ❌ "Redis seria cool para distributed workflows"
  ❌ "Um dia talvez preciso de S3 loading"
```

#### **Teste 3: Arquitetura está pronta?**

```
Pergunta: "Posso implementar em <2 dias sem refatoração?"

Pronto (go ahead):
  ✅ Repository interface é genérica
  ✅ TypedRepository[T] é agnóstico de source
  ✅ Service layer não depende de FS specifics
  ✅ Novo adapter é apenas outra impl de Repository

Não pronto (refatore primeiro):
  ❌ Service tem lógica hardcoded de FS/Git
  ❌ Watcher é específico de fsnotify
  ❌ Transactionality assume Git
```

## Exemplos de Aplicação

### Exemplo 1: Arbour HTTP Adapter (Hypothetical Future)

**Scenario**: Arbour Phase 1 está em produção. Users instalam flows:

```bash
arbour install flow:insurance-workflow
```

**Questão**: Precisa loam2http adapter?

- **Test 1**: "Cliente pode fazer HTTP.Get(...) e passar?"
  - Resposta: Sim, mas recarregar sem restart é difícil
  - Teste 1 = **Talvez Sim**, merece adapter se usar Loam.Watch()

- **Test 2**: "Há pressão?"
  - Resposta: Users reclamando "reload lento"? Sim → **Sim**
  - Resposta: Está funcionando OK? Não → **Não**

- **Test 3**: "Arquitetura pronta?"
  - Resposta: HTTPRepository impl 2 dias? Sim → **Sim**
  - Resposta: Precisa refator Watch logic? Talvez → **Talvez**

**Decisão**: Se Tests 1 & 2 são "Sim" e Test 3 é "Sim" → Implementa

### Exemplo 2: Loam2Database Adapter (Hypothetical)

**Scenario**: Alguém pede "Loam deveria carregar workflows de Postgres"

**Questão**: Deveria ser adapter de Loam?

- **Test 1**: "É responsabilidade de Loam?"
  - Pergunta: "Ou é de DALgo?"
  - Resposta: Se múltiplos frameworks (Trellis, Life-DSL, Custom) precisam abstração agnóstica → **Talvez Loam**
  - Resposta: Se é uma aplicação específica → **Não, use DALgo ou seu código**
  - Resultado: **Provavelmente Não** (DALgo é melhor fit)

- **Test 2**: "Há pressão?"
  - Resposta: Ninguém pediu ainda → **Não**

**Decisão**: Não implementa. Redireciona para DALgo.

### Exemplo 3: Life-DSL Workers (Current Phase 2b)

**Scenario**: Life-DSL Phase é iniciada. Workers são definidos em arquivos.

**Questão**: Precisa novo adapter para workers?

- **Test 1**: "É responsabilidade de Loam?"
  - Resposta: Workers são "Documentos com Metadados" como Trellis nodes
  - Resultado: **Sim, mesmo padrão de FS adapter**

- **Test 2**: "Há pressão?"
  - Resposta: Life-DSL depende disso para funcionar
  - Resultado: **Sim, pressão direta**

- **Test 3**: "Arquitetura pronta?"
  - Resposta: FS adapter funciona para workers também
  - Resultado: **Sim, usa existente, sem novo adapter**

**Decisão**: Não precisa novo adapter. FS adapter basta.

## Roadmap de Adapters (Priorizado por Demanda Real)

| Adapter | Need Now | Strong Signal | ETA | Why |
|---------|----------|---|-----|-----|
| **FS + Git** | ✅ Critical | ✅ | v0.10 ✓ | Trellis + Ecosystem |
| **HTTP** | ❌ No | ⏳ Arbour P1 | v0.12+ | Only if Arbour shows need |
| **Redis** | ❌ No | ⏳ Life-DSL distributed | v0.13+ | Only if scaling distributed |
| **Database** | ❌ No | ❌ Never | N/A | Use DALgo instead |
| **S3** | ❌ No | ⏳ SaaS Trellis | Future | Only if cloud variant |
| **Memory** | ❌ No | ⏳ Testing | v0.12 | Simple, maybe for tests |

## Implementação Futura (Quando Chegar a Hora)

Se a demanda apertar para um novo adapter:

1. **Criar novo arquivo**: `pkg/adapters/httpadapter/http.go`
2. **Implementar Repository interface**:

   ```go
   type HTTPRepository struct {
       baseURL string
       cache   map[string]*loam.Document
   }
   
   func (h *HTTPRepository) Get(ctx context.Context, id string) (*Document, error) {
       // HTTP call + caching logic
   }
   
   func (h *HTTPRepository) Watch(ctx context.Context) (<-chan EventStream, error) {
       // Polling or webhook logic
   }
   ```

3. **Write tests** contra case de uso real
4. **Document** exatamente when/why usar esse adapter
5. **Add examples** mostrando integração

## Consequências

✅ **Mantém Loam focado**: Não bloated com adapters não usados  
✅ **Simples de manter**: Menos código, menos bugs  
✅ **Pronto para escalar**: Quando demanda chegar, implementa rápido  
❌ **Risco teórico**: Se múltipls adapters forem précis no mesmo tempo, refactoring pode ser necessário

## Notas

- Veja [ADR 014: Trellis as Primary Stakeholder](014-understanding-trellis-loam-symbiosis.md) para entender por quê FS+Git é especialização correta
- Veja [PLANNING.md Phase 0.11](../PLANNING.md#fase-0110-ecosystem-positioning) para roadmap detalhado
