package db

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var ErrKeyNotFound = errors.New("chave não encontrada")

// KVPair é a struct usada pelo rbtree.go e sstable.go
type KVPair struct {
	Key   []byte
	Value []byte
}

const MemTableLimit = 4096

// Engine agora controla a memória Ativa, a Imutável e a fila de disco
type Engine struct {
	mu            sync.RWMutex
	activeTree    *RBTree
	immutableTree *RBTree
	flushCh       chan *RBTree
}

// NewEngine inicializa o motor de armazenamento
func NewEngine() *Engine {
	e := &Engine{
		activeTree: NewRBTree(),
		flushCh:    make(chan *RBTree, 1),
	}
	go e.flushWorker()
	return e
}

// flushWorker salva as árvores congeladas no disco
func (e *Engine) flushWorker() {
	fileIndex := 1
	for treeToFlush := range e.flushCh {
		fmt.Printf("[Worker] Iniciando flush de %d chaves...\n", treeToFlush.Size)

		orderedData := treeToFlush.GetAllInOrder()
		filename := fmt.Sprintf("sstable_%04d.data", fileIndex)

		err := WriteSSTable(orderedData, filename)
		if err != nil {
			fmt.Printf("[Worker - ERRO CRÍTICO] Falha ao salvar %s: %v\n", filename, err)
		} else {
			fmt.Printf("[Worker] %s salvo e encriptado com sucesso!\n", filename)
			fileIndex++
		}

		// Libera a árvore imutável da RAM
		e.mu.Lock()
		e.immutableTree = nil
		e.mu.Unlock()
	}
}

// Put roteia a escrita e verifica limites
func (e *Engine) Put(key []byte, value []byte) error {
	e.mu.Lock()

	if e.activeTree.Size >= MemTableLimit {
		if e.immutableTree != nil {
			e.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return e.Put(key, value)
		}

		fmt.Println("[Maestro] Limite atingido. Congelando MemTable ativa...")
		e.immutableTree = e.activeTree
		e.activeTree = NewRBTree()
		e.flushCh <- e.immutableTree
	}

	currentActive := e.activeTree
	e.mu.Unlock()

	return currentActive.Put(key, value)
}

// Get lê primeiro da ativa, depois da imutável
func (e *Engine) Get(key []byte) ([]byte, error) {
	e.mu.RLock()
	active := e.activeTree
	immutable := e.immutableTree
	e.mu.RUnlock()

	// 1. Busca na MemTable Ativa (Tempo real)
	val, err := active.Get(key)
	if err == nil {
		return val, nil
	}

	// 2. Busca na Imutável (Aguardando o worker de disco)
	if immutable != nil {
		val, err = immutable.Get(key)
		if err == nil {
			return val, nil
		}
	}

	// 3. Busca no Disco (SSTables Encriptadas)
	// Lista todos os arquivos que batem com o padrão sstable_*.data
	files, err := filepath.Glob("sstable_*.data")
	if err == nil && len(files) > 0 {
		// Ordena os arquivos em ordem reversa (ex: sstable_0003, depois 0002, depois 0001)
		// É vital para garantir que pegamos a versão mais atualizada da chave!
		sort.Sort(sort.Reverse(sort.StringSlice(files)))

		for _, file := range files {
			val, err := FindInSSTable(key, file)
			if err == nil {
				// Achou no disco! Retorna com sucesso.
				return val, nil
			}
		}
	}

	// Se chegou aqui, a chave realmente não existe em lugar nenhum do banco.
	return nil, ErrKeyNotFound
}

func (e *Engine) Delete(key []byte) error {
	return nil
}
