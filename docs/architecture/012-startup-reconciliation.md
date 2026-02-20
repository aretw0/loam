# ADR 012: Startup Reconciliation (Cold Start)

**Data:** Historicamente Adotada (Registrada em Fev/2026)
**Status:** Aceita

## Contexto

Um aplicativo reativo gerindo estado local-first sofre de um problema clássico: ele só sabe das mudanças que presenciou enquanto estava ligado.
Se a aplicação cliente fechar, o desenvolvedor editar um `nota_x.md` ou excluir o `foto_2.md` via gerenciador de arquivos do Windows/Mac, e no dia seguinte a aplicação for reaberta: a memória que o Loam reconstruiria do Cache Local (`index.json`) exibiria a `foto_2.md` como existindo e a `nota_x` desatualizada, desincronizando a interface (Stale State).

## Decisão

Projetou-se um fluxo arquitetural de **Cold Start Repair via "Visited Maps"** que atua sequencial e imperativamente *antes* que a esteira de `Watch(OS)` comece a rodar.

1. O Serviço chama e aloca o último Cache salvo mapeando entradas e populando-as num Map em que o valor padrão associado a todos seja `[Visited = False]`.
2. O Adapter inicializa o varrimento de Disco Local (`Walk Filesystem`) e compara com a tabela:
    - O arquivo não existia no Cache → (Novo Documento detectado, evento emitido como **Create** pós-arranque).
    - O arquivo existe e possui Diff de `mtime` temporal → (Documento Atualizado externamente, emite **Modify**).
    - Ambas situações a entry sofre "check in": `[Visited = True]`.
3. Terminado o Walk, iteramos sobre todo o resto da tabela do Visited Map original.
    - Aquilo que manteve-se no status de inicialização `[Visited = False]` foi varrido da face do disco mas o Loam lembra. Trata-se de uma exclusão externa ao app. (Emite-se um evento **Delete**).
4. O Index assincronamente re-escreve a si mesmo (self-heals) e só então o Watcher assume a gestão de tempo em "tempo-real".

## Consequências

- **Pró (Auto-Cura):** Um sistema Loam pode sofrer pull requests agressivos injetando 5.000 alterações pelo GitHub (sem um listener) e assim que a aplicação abrir, 5.000 eventos de tela/socket serão disparados, ajeitando o client.
- **Contra (Overhead de Boot Lento):** Para Vaults com milhões de pequenas instâncias (ex: um cluster node modules dump), a inicialização sofre consideravelmente pelo limite de `I/O Seek` imposto por realizar o Walk inicial contra o mapa em memória. Dependendo do SSD pode acarretar engasgos visíveis na abertura de "Aplicações de Interface Síncrona" (como Splash Screens).
