package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// Configuração do Spike
const (
	NumFiles    = 100
	WorkerCount = 100
)

// Global lock para operações Git (simulando o gerenciador de transações)
var gitLock sync.Mutex

func main() {
	log.Println("⚡ Iniciando Spike: Loam Concurency Test")

	// 1. Setup do Diretório Temporário
	tmpDir, err := os.MkdirTemp("", "loam-spike-*")
	if err != nil {
		log.Fatalf("Erro ao criar temp dir: %v", err)
	}
	// Limpeza no final (comentado para inspeção se falhar, descomentar para produção)
	// defer os.RemoveAll(tmpDir)

	log.Printf("📂 Diretório de trabalho: %s", tmpDir)

	// 2. Inicializar Git
	runGit(tmpDir, "init")
	// Configurar user dummy para o commit funcionar
	runGit(tmpDir, "config", "user.email", "spike@loam.dev")
	runGit(tmpDir, "config", "user.name", "Loam Spike")

	start := time.Now()

	// 3. Execução Concorrente
	var wg sync.WaitGroup
	wg.Add(WorkerCount)

	for i := 0; i < WorkerCount; i++ {
		go func(id int) {
			defer wg.Done()
			filename := fmt.Sprintf("note_%d.md", id)
			content := fmt.Sprintf("---\nid: %d\ntimestamp: %s\n---\n# Nota %d\nConteúdo de teste para o spike.", id, time.Now().Format(time.RFC3339), id)

			// Simula operação de IO (escrita no disco)
			err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
			if err != nil {
				log.Printf("[Erro] Falha na escrita %s: %v", filename, err)
				return
			}

			// Seção Crítica: Commit
			// O Git não suporta múltiplos processos mexendo no index/lock ao mesmo tempo
			gitLock.Lock()
			defer gitLock.Unlock()

			// Adiciona apenas este arquivo
			if err := runGit(tmpDir, "add", filename); err != nil {
				log.Printf("[Erro] Git Add %s: %v", filename, err)
				return
			}

			// Commit
			if err := runGit(tmpDir, "commit", "-m", fmt.Sprintf("chore: update %s", filename)); err != nil {
				log.Printf("[Erro] Git Commit %s: %v", filename, err)
				return
			}

			// log.Printf("✅ Commit %d ok", id) // Comentado para reduzir ruído
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// 4. Validação Final
	log.Println("🏁 Todas as goroutines finalizaram.")
	log.Printf("⏱️  Tempo Total: %v", duration)
	log.Printf("⚡ Throughput: %.2f commits/seg", float64(NumFiles)/duration.Seconds())

	// Verificar git status
	status := getGitOutput(tmpDir, "status", "--porcelain")
	if status != "" {
		log.Fatalf("❌ FALHA: Git status não está limpo:\n%s", status)
	} else {
		log.Println("✅ SUCESSO: Git status limpo (clean slate).")
	}

	// Contar commits
	count := getGitOutput(tmpDir, "rev-list", "--count", "HEAD")
	log.Printf("📊 Total de Commits no Histórico: %s", count)
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v failed: %v\nOutput: %s", args, err, string(out))
	}
	return nil
}

func getGitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Erro ao ler status: %v", err)
		return ""
	}
	return string(out)
}
