package db

import (
	"bytes"
	"os"
	"testing"
)

func TestSSTable_WriteAndRead(t *testing.T) {
	filename := "test_sstable_9999.data"

	// Limpa o arquivo de teste antes e depois de rodar
	os.Remove(filename)
	defer os.Remove(filename)

	// 1. Prepara dados falsos (simulando a saída do inOrderTraversal)
	dados := []KVPair{
		{Key: []byte("a_chave"), Value: []byte("valor_a")},
		{Key: []byte("b_chave"), Value: []byte("valor_b")},
		{Key: []byte("c_chave"), Value: []byte("valor_c")},
	}

	// 2. Escreve no disco (encriptado e empacotado em binário)
	err := WriteSSTable(dados, filename)
	if err != nil {
		t.Fatalf("Falha ao escrever SSTable: %v", err)
	}

	// 3. Lê do disco buscando uma chave existente
	val, err := FindInSSTable([]byte("b_chave"), filename)
	if err != nil {
		t.Fatalf("Erro ao buscar chave existente: %v", err)
	}

	if !bytes.Equal(val, []byte("valor_b")) {
		t.Errorf("Esperava 'valor_b', recebeu '%s'", string(val))
	}

	// 4. Busca uma chave inexistente (deve retornar ErrKeyNotFound)
	_, err = FindInSSTable([]byte("z_chave_fantasma"), filename)
	if err != ErrKeyNotFound {
		t.Errorf("Esperava ErrKeyNotFound para chave inexistente, recebeu %v", err)
	}
}
