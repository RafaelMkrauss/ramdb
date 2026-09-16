package db

import "errors"

var ErrKeyNotFound = errors.New("chave não encontrada")

// Engine é o ponto de entrada do motor de armazenamento.
type Engine struct {
	Tree *RBTree // A sua Árvore Rubro-Negra é o motor principal
}

// NewEngine inicializa o motor instanciando a árvore
func NewEngine() *Engine {
	return &Engine{
		Tree: NewRBTree(),
	}
}

// Put insere os bytes diretamente na árvore
func (e *Engine) Put(key []byte, value []byte) error {
	return e.Tree.Put(key, value)
}

// Get busca os bytes diretamente na árvore
func (e *Engine) Get(key []byte) ([]byte, error) {
	return e.Tree.Get(key)
}

// Delete remove a chave (deixaremos o gancho pronto para quando você implementar a lógica de deleção na árvore)
func (e *Engine) Delete(key []byte) error {
	// e.Tree.Delete(key) <- A ser implementado na RBTree futuramente
	return nil
}
