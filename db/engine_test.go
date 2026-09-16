package db

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEngine_Lifecycle_And_Flush(t *testing.T) {
	// Limpa qualquer sstable residual gerada por execuções anteriores
	limparSSTables()
	defer limparSSTables()

	engine := NewEngine()

	// O limite do motor (MemTableLimit) é 4096.
	// Vamos inserir 4100 chaves para forçar EXATAMENTE 1 flush para o disco.
	for i := 0; i < 4100; i++ {
		key := []byte(fmt.Sprintf("chave_%04d", i))
		val := []byte(fmt.Sprintf("valor_%04d", i))
		err := engine.Put(key, val)
		if err != nil {
			t.Fatalf("Erro no Put da iteração %d: %v", i, err)
		}
	}

	// Aguarda 1 segundo para a goroutine do worker assíncrono terminar de gravar no HD
	time.Sleep(1 * time.Second)

	// --- FASE DE LEITURA ---

	// 1. Lendo da MemTable Ativa (As chaves mais recentes ficaram na RAM: 4096 a 4099)
	valMem, err := engine.Get([]byte("chave_4098"))
	if err != nil || !bytes.Equal(valMem, []byte("valor_4098")) {
		t.Errorf("Falha ao ler chave que deveria estar na MemTable Ativa: %v", err)
	}

	// 2. Lendo do Disco (As chaves mais antigas foram pro arquivo sstable_0001.data)
	valDisco, err := engine.Get([]byte("chave_0500"))
	if err != nil || !bytes.Equal(valDisco, []byte("valor_0500")) {
		t.Errorf("Falha ao ler chave que deveria estar no Disco: %v", err)
	}

	// 3. Lendo chave que não existe em nenhum lugar
	_, err = engine.Get([]byte("chave_9999"))
	if err != ErrKeyNotFound {
		t.Errorf("Esperava ErrKeyNotFound, recebeu: %v", err)
	}
}

// limparSSTables é uma função auxiliar para não sujar o seu repositório com lixo de testes
func limparSSTables() {
	files, _ := filepath.Glob("sstable_*.data")
	for _, f := range files {
		os.Remove(f)
	}
}
