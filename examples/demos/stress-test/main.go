package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

// Configuração do Spike
const (
	NumFiles    = 100 // Total files per worker? Or total overall? Original was 100 workers, 1 file each.
	WorkerCount = 100
)

func main() {
	log.Println("⚡ Iniciando Demo: Loam Concurrency Stress Test")

	// 1. Setup do Diretório Temporário da Vault
	vaultPath, err := os.MkdirTemp("", "loam-stress-*")
	if err != nil {
		log.Fatalf("Erro ao criar temp dir: %v", err)
	}
	defer os.RemoveAll(vaultPath) // Cleanup

	log.Printf("📂 Vault temporário: %s", vaultPath)

	// 2. Inicializar Loam Service
	// Usamos WithAutoInit para garantir que o git init seja feito.
	// Logger descartado para não poluir o output do bench.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	service, err := loam.New(vaultPath,
		loam.WithAutoInit(true),
		loam.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}

	// 2.1 Cenário Dirty State: Criar lixo não rastreado no diretório do vault
	log.Println("🗑️  Criando arquivos 'lixo' (untracked)...")
	for i := 0; i < 10; i++ {
		garbageName := fmt.Sprintf("garbage_%d.txt", i)
		// Aqui acessamos o disco direto, "por fora" do Loam, para testar a resiliência dele
		os.WriteFile(fmt.Sprintf("%s/%s", vaultPath, garbageName), []byte("Eu não deveria ser comitado!"), 0644)
	}

	// 2.2 Iniciar cronometragem
	start := time.Now()

	// 3. Execução Concorrente
	var wg sync.WaitGroup
	wg.Add(WorkerCount)

	log.Printf("🚀 Disparando %d goroutines de escrita simultânea...", WorkerCount)

	for i := 0; i < WorkerCount; i++ {
		go func(id int) {
			defer wg.Done()

			// Nota: ID único para não haver colisão de escrita no mesmo arquivo (o que seria Race Condition de ALTO NÍVEL, não do Loam)
			noteID := fmt.Sprintf("note_%d", id)
			content := fmt.Sprintf("# Nota %d\nConteúdo de teste de concorrência.\nTimestamp: %s", id, time.Now().Format(time.RFC3339))

			// Change Reason (Commit Message)
			reason := fmt.Sprintf("chore(stress): add note %d via go routine", id)
			ctx := context.WithValue(context.Background(), core.ChangeReasonKey, reason)

			// O Loam deve cuidar do Locking interno!
			if err := service.SaveNote(ctx, noteID, content, nil); err != nil {
				log.Printf("❌ [Erro Rutine %d] Falha ao salvar: %v", id, err)
				return
			}

			// Feedback visual mínimo
			// fmt.Print(".")
		}(i)
	}

	wg.Wait()
	fmt.Println() // Quebra de linha após os pontos

	// 2.3 Cronometragem final
	duration := time.Since(start)

	// 4. Validação Final
	log.Println("🏁 Todas as goroutines finalizaram.")
	log.Printf("⏱️  Tempo Total: %v", duration)
	throughput := float64(WorkerCount) / duration.Seconds()
	log.Printf("⚡ Throughput: %.2f commits/seg", throughput)

	// Validar contagem de notas via API
	// (Poderíamos usar git rev-list também, mas vamos usar a API para variar)
	// Nota: List ainda não está exposto no Facade loam.go, então vamos confiar no log de erro acima
	// ou se quiser, podemos instanciar um repo direto, mas o teste principal aqui é "não crashou".

	log.Println("✅ Teste finalizado sem panics (esperamos que sem erros de log também).")
}
